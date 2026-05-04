package table

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("not the restaurant owner")
)

type restaurantRepo interface {
	FindOwnerID(ctx context.Context, id string) (string, error)
}

type Service struct {
	repo           *Repository
	restaurantRepo restaurantRepo
}

func NewService(repo *Repository, restaurantRepo restaurantRepo) *Service {
	return &Service{repo: repo, restaurantRepo: restaurantRepo}
}

func (s *Service) Create(ctx context.Context, restaurantID, userID string, req CreateReq) (*Response, error) {
	if err := s.checkOwnership(ctx, restaurantID, userID); err != nil {
		return nil, err
	}

	t := &Table{
		RestaurantID: restaurantID,
		Number:       req.Number,
		Capacity:     req.Capacity,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return toResponse(t), nil
}

func (s *Service) GetByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	tables, err := s.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(tables))
	for i, t := range tables {
		responses[i] = *toResponse(&t)
	}
	return responses, nil
}

func (s *Service) UpdateCapacity(ctx context.Context, restaurantID, tableID, userID string, req UpdateReq) (*Response, error) {
	if err := s.checkOwnership(ctx, restaurantID, userID); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateCapacity(ctx, tableID, req.Capacity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update capacity: %w", err)
	}

	t, err := s.repo.FindByID(ctx, tableID)
	if err != nil {
		return nil, err
	}
	return toResponse(t), nil
}

func (s *Service) Delete(ctx context.Context, restaurantID, tableID, userID string) error {
	if err := s.checkOwnership(ctx, restaurantID, userID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, tableID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("delete table: %w", err)
	}
	return nil
}

func (s *Service) checkOwnership(ctx context.Context, restaurantID, userID string) error {
	ownerID, err := s.restaurantRepo.FindOwnerID(ctx, restaurantID)
	if err != nil {
		return fmt.Errorf("find restaurant: %w", err)
	}
	if ownerID == "" {
		return ErrNotFound
	}
	if ownerID != userID {
		return ErrForbidden
	}
	return nil
}

func toResponse(t *Table) *Response {
	return &Response{
		ID:           t.ID,
		RestaurantID: t.RestaurantID,
		Number:       t.Number,
		Capacity:     t.Capacity,
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
