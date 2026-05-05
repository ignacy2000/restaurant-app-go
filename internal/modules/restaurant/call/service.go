package call

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"table-service.pl/internal/modules/restaurant/table"
	"table-service.pl/pkg/ws"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	repo      *Repository
	tableRepo *table.Repository
	hub       *ws.Hub
}

func NewService(repo *Repository, tableRepo *table.Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, tableRepo: tableRepo, hub: hub}
}

func (s *Service) Create(ctx context.Context, restaurantID, tableID string) (*Response, error) {
	t, err := s.tableRepo.FindByID(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("find table: %w", err)
	}
	if t == nil || t.RestaurantID != restaurantID {
		return nil, ErrNotFound
	}

	c := &WaiterCall{
		RestaurantID: restaurantID,
		TableID:      tableID,
		Status:       StatusPending,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create call: %w", err)
	}

	resp := toResponse(c)
	s.hub.BroadcastToRestaurant(restaurantID, ws.Event{
		Type:    ws.EventCallCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *Service) GetActiveByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	calls, err := s.repo.FindActiveByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(calls))
	for i, c := range calls {
		responses[i] = *toResponse(&c)
	}
	return responses, nil
}

func (s *Service) GetByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	calls, err := s.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(calls))
	for i, c := range calls {
		responses[i] = *toResponse(&c)
	}
	return responses, nil
}

func (s *Service) UpdateStatus(ctx context.Context, callID string, req UpdateStatusReq) error {
	call, err := s.repo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("find call: %w", err)
	}
	if call == nil {
		return ErrNotFound
	}

	if err := s.repo.UpdateStatus(ctx, callID, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("update status: %w", err)
	}

	call.Status = req.Status
	s.hub.BroadcastToRestaurantAndTable(call.RestaurantID, call.TableID, ws.Event{
		Type:    ws.EventCallStatusChanged,
		Payload: toResponse(call),
	})
	return nil
}

func toResponse(c *WaiterCall) *Response {
	return &Response{
		ID:           c.ID,
		RestaurantID: c.RestaurantID,
		TableID:      c.TableID,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
