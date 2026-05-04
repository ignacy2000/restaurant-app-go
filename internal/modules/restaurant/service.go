package restaurant

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateReq) (*Response, error) {
	rest := &Restaurant{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
	}
	if err := s.repo.Create(ctx, rest); err != nil {
		return nil, fmt.Errorf("create restaurant: %w", err)
	}
	return toResponse(rest), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Response, error) {
	rest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rest == nil {
		return nil, nil
	}
	return toResponse(rest), nil
}

func (s *Service) GetByUser(ctx context.Context, userID string) ([]Response, error) {
	restaurants, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(restaurants))
	for i, r := range restaurants {
		responses[i] = *toResponse(&r)
	}
	return responses, nil
}

func toResponse(r *Restaurant) *Response {
	return &Response{
		ID:          r.ID,
		UserID:      r.UserID,
		Name:        r.Name,
		Description: r.Description,
		Address:     r.Address,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
