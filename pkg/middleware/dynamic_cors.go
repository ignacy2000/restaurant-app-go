package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// corsOriginsKey is the Redis Set key that stores allowed origins.
// A sentinel member marks the set as populated (handles empty-set case).
const corsOriginsKey = "cors:allowed_origins"
const corsLoadedSentinel = "\x00loaded\x00"

type OriginStore struct {
	load   func(ctx context.Context) ([]string, error)
	static []string
	ttl    time.Duration
	rdb    *redis.Client
}

func NewOriginStore(ttl time.Duration, load func(context.Context) ([]string, error), rdb *redis.Client, static ...string) *OriginStore {
	return &OriginStore{load: load, static: static, ttl: ttl, rdb: rdb}
}

func (s *OriginStore) allowed(ctx context.Context, origin string) bool {
	pipe := s.rdb.Pipeline()
	isLoaded := pipe.SIsMember(ctx, corsOriginsKey, corsLoadedSentinel)
	isMember := pipe.SIsMember(ctx, corsOriginsKey, origin)
	pipe.Exec(ctx) //nolint:errcheck

	if loaded, _ := isLoaded.Result(); !loaded {
		if err := s.refresh(ctx); err != nil {
			return false
		}
		ok, err := s.rdb.SIsMember(ctx, corsOriginsKey, origin).Result()
		return err == nil && ok
	}

	ok, _ := isMember.Result()
	return ok
}

func (s *OriginStore) refresh(ctx context.Context) error {
	dbOrigins, err := s.load(ctx)
	if err != nil {
		return err
	}

	members := make([]interface{}, 0, len(dbOrigins)+len(s.static)+1)
	members = append(members, corsLoadedSentinel)
	for _, o := range dbOrigins {
		members = append(members, o)
	}
	for _, o := range s.static {
		members = append(members, o)
	}

	pipe := s.rdb.TxPipeline()
	pipe.Del(ctx, corsOriginsKey)
	pipe.SAdd(ctx, corsOriginsKey, members...)
	pipe.Expire(ctx, corsOriginsKey, s.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

// DynamicCORS zastępuje middleware.CORS() — sprawdza Origin requestu
// przeciwko liście z bazy danych (cache w Redis, odświeżany co TTL) i domenie z .env.
// Requesty bez nagłówka Origin (np. curl, server-to-server) są przepuszczane.
func DynamicCORS(store *OriginStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin == "" {
			c.Next()
			return
		}

		if !store.allowed(c.Request.Context(), origin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
