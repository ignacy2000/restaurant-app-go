package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// OriginStore trzyma dozwolone originy w pamięci i odświeża je z DB co TTL.
// Użyj NewOriginStore i przekaż do DynamicCORS.
type OriginStore struct {
	load   func(ctx context.Context) ([]string, error)
	static []string
	ttl    time.Duration

	mu       sync.RWMutex
	cache    map[string]struct{}
	loadedAt time.Time
}

func NewOriginStore(ttl time.Duration, load func(context.Context) ([]string, error), static ...string) *OriginStore {
	return &OriginStore{
		load:   load,
		static: static,
		ttl:    ttl,
		cache:  make(map[string]struct{}),
	}
}

func (s *OriginStore) allowed(origin string) bool {
	s.mu.RLock()
	fresh := time.Since(s.loadedAt) < s.ttl
	_, ok := s.cache[origin]
	s.mu.RUnlock()

	if fresh {
		return ok
	}
	s.refresh()

	s.mu.RLock()
	_, ok = s.cache[origin]
	s.mu.RUnlock()
	return ok
}

func (s *OriginStore) refresh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.loadedAt) < s.ttl {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbOrigins, err := s.load(ctx)
	if err != nil {
		return
	}

	next := make(map[string]struct{}, len(dbOrigins)+len(s.static))
	for _, o := range dbOrigins {
		next[o] = struct{}{}
	}
	for _, o := range s.static {
		next[o] = struct{}{}
	}
	s.cache = next
	s.loadedAt = time.Now()
}

// DynamicCORS zastępuje middleware.CORS() — sprawdza Origin requestu
// przeciwko liście z bazy danych (odświeżanej co TTL) i domenie z .env.
// Requesty bez nagłówka Origin (np. curl, server-to-server) są przepuszczane.
func DynamicCORS(store *OriginStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin == "" {
			c.Next()
			return
		}

		if !store.allowed(origin) {
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
