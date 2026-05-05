package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
	"table-service.pl/internal/modules/user"
	"table-service.pl/pkg/worker"
)

type LoginReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Service struct {
	repo        *Repository
	userRepo    *user.Repository
	jwtSecret   string
	asynqClient *asynq.Client
	frontendURL string
}

func NewService(repo *Repository, userRepo *user.Repository, jwtSecret string, asynqClient *asynq.Client, frontendURL string) *Service {
	return &Service{repo: repo, userRepo: userRepo, jwtSecret: jwtSecret, asynqClient: asynqClient, frontendURL: frontendURL}
}

func (s *Service) Login(ctx context.Context, req LoginReq) (*TokenPair, error) {
	u, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.newAccessToken(u.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		UserID:       u.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteByRefreshToken(ctx, refreshToken)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	session, err := s.repo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if err := s.repo.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("delete old session: %w", err)
	}

	accessToken, err := s.newAccessToken(session.UserID)
	if err != nil {
		return nil, err
	}

	newRT, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	newSession := &Session{
		UserID:       session.UserID,
		RefreshToken: newRT,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.repo.CreateSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: newRT}, nil
}

func (s *Service) newAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Service) ListOrigins(ctx context.Context) ([]AllowedOrigin, error) {
	return s.repo.ListOrigins(ctx)
}

func (s *Service) AddOrigin(ctx context.Context, origin string) (*AllowedOrigin, error) {
	return s.repo.AddOrigin(ctx, origin)
}

func (s *Service) RemoveOrigin(ctx context.Context, id string) error {
	return s.repo.RemoveOrigin(ctx, id)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return nil
	}

	token, err := newRefreshToken()
	if err != nil {
		return err
	}

	rt := &PasswordResetToken{
		UserID:    u.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.repo.CreateResetToken(ctx, rt); err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}

	resetURL := s.frontendURL + "/reset-password?token=" + token
	task, err := worker.NewPasswordResetTask(email, resetURL)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	if _, err := s.asynqClient.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	rt, err := s.repo.FindResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("find reset token: %w", err)
	}
	if rt == nil || rt.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("invalid or expired token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, rt.UserID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return s.repo.DeleteResetToken(ctx, token)
}
