package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"table-service.pl/internal/modules/restaurant/table"
	"table-service.pl/pkg/ws"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("not the restaurant owner")
)

type Service struct {
	repo      *Repository
	tableRepo *table.Repository
	hub       *ws.Hub
}

func NewService(repo *Repository, tableRepo *table.Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, tableRepo: tableRepo, hub: hub}
}

func (s *Service) Create(ctx context.Context, restaurantID, tableID string, req CreateReq) (*Response, error) {
	t, err := s.tableRepo.FindByID(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("find table: %w", err)
	}
	if t == nil || t.RestaurantID != restaurantID {
		return nil, ErrNotFound
	}

	o := &Order{
		RestaurantID: restaurantID,
		TableID:      tableID,
		Status:       StatusPending,
		Notes:        req.Notes,
		Items:        make([]OrderItem, len(req.Items)),
	}
	for i, it := range req.Items {
		o.Items[i] = OrderItem{Name: it.Name, Quantity: it.Quantity, Notes: it.Notes}
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	resp := toResponse(o)
	s.hub.BroadcastToRestaurant(restaurantID, ws.Event{
		Type:    ws.EventOrderCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *Service) GetByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	orders, err := s.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(orders))
	for i, o := range orders {
		responses[i] = *toResponse(&o)
	}
	return responses, nil
}

func (s *Service) UpdateStatus(ctx context.Context, orderID string, req UpdateStatusReq) error {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}
	if o == nil {
		return ErrNotFound
	}

	if err := s.repo.UpdateStatus(ctx, orderID, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("update status: %w", err)
	}

	o.Status = req.Status
	s.hub.BroadcastToRestaurantAndTable(o.RestaurantID, o.TableID, ws.Event{
		Type:    ws.EventOrderStatusChanged,
		Payload: toResponse(o),
	})
	return nil
}

func toResponse(o *Order) *Response {
	items := make([]ItemResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemResponse{ID: it.ID, Name: it.Name, Quantity: it.Quantity, Notes: it.Notes}
	}
	return &Response{
		ID:           o.ID,
		RestaurantID: o.RestaurantID,
		TableID:      o.TableID,
		Status:       o.Status,
		Notes:        o.Notes,
		Items:        items,
		CreatedAt:    o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
