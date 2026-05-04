package menu

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("restaurant not found")
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
	ownerID, err := s.restaurantRepo.FindOwnerID(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find restaurant: %w", err)
	}
	if ownerID == "" {
		return nil, ErrNotFound
	}
	if ownerID != userID {
		return nil, ErrForbidden
	}

	m := &Menu{
		RestaurantID: restaurantID,
		Name:         req.Name,
		Description:  req.Description,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return toResponse(m), nil
}

func (s *Service) GetByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	menus, err := s.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(menus))
	for i, m := range menus {
		responses[i] = *toResponse(&m)
	}
	return responses, nil
}

func (s *Service) CreateItem(ctx context.Context, menuID, userID string, req CreateItemReq) (*ItemResponse, error) {
	restaurantID, err := s.repo.FindMenuRestaurantID(ctx, menuID)
	if err != nil {
		return nil, fmt.Errorf("find menu: %w", err)
	}
	if restaurantID == "" {
		return nil, ErrNotFound
	}
	ownerID, err := s.restaurantRepo.FindOwnerID(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find owner: %w", err)
	}
	if ownerID != userID {
		return nil, ErrForbidden
	}
	item := &MenuItem{
		MenuID:      menuID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Position:    req.Position,
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	return toItemResponse(item), nil
}

func (s *Service) GetItemsByRestaurant(ctx context.Context, restaurantID string) ([]ItemResponse, error) {
	items, err := s.repo.FindItemsByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]ItemResponse, len(items))
	for i, item := range items {
		responses[i] = *toItemResponse(&item)
	}
	return responses, nil
}

func (s *Service) DeleteItem(ctx context.Context, itemID, userID string) error {
	restaurantID, err := s.repo.FindItemRestaurantID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("find item: %w", err)
	}
	if restaurantID == "" {
		return ErrNotFound
	}
	ownerID, err := s.restaurantRepo.FindOwnerID(ctx, restaurantID)
	if err != nil {
		return fmt.Errorf("find owner: %w", err)
	}
	if ownerID != userID {
		return ErrForbidden
	}
	return s.repo.DeleteItem(ctx, itemID)
}

func toResponse(m *Menu) *Response {
	return &Response{
		ID:           m.ID,
		RestaurantID: m.RestaurantID,
		Name:         m.Name,
		Description:  m.Description,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toItemResponse(item *MenuItem) *ItemResponse {
	return &ItemResponse{
		ID:          item.ID,
		MenuID:      item.MenuID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Position:    item.Position,
	}
}
