package order

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"table-service.pl/internal/modules/restaurant/table"
	"table-service.pl/pkg/mailer"
	"table-service.pl/pkg/ws"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("not the restaurant owner")
)

type restaurantNamer interface {
	FindName(ctx context.Context, id string) (string, error)
}

type Service struct {
	repo           *Repository
	tableRepo      *table.Repository
	hub            *ws.Hub
	restaurantRepo restaurantNamer
	mailer         *mailer.Mailer
	frontendURL    string
}

func NewService(repo *Repository, tableRepo *table.Repository, hub *ws.Hub, restaurantRepo restaurantNamer, mail *mailer.Mailer, frontendURL string) *Service {
	return &Service{repo: repo, tableRepo: tableRepo, hub: hub, restaurantRepo: restaurantRepo, mailer: mail, frontendURL: frontendURL}
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
		Status:       StatusAwaitingConfirmation,
		Notes:        req.Notes,
		GuestEmail:   req.GuestEmail,
		Items:        make([]OrderItem, len(req.Items)),
	}
	for i, it := range req.Items {
		o.Items[i] = OrderItem{Name: it.Name, Quantity: it.Quantity, Notes: it.Notes}
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	token, err := newToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	conf := &OrderConfirmation{
		OrderID:   o.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.repo.CreateConfirmation(ctx, conf); err != nil {
		return nil, fmt.Errorf("create confirmation: %w", err)
	}

	if s.mailer != nil {
		tableNumber := t.Number
		items := o.Items
		notes := req.Notes
		email := req.GuestEmail
		confirmURL := s.frontendURL + "/order-confirmed?token=" + token
		go func() {
			restaurantName := ""
			if s.restaurantRepo != nil {
				if name, e := s.restaurantRepo.FindName(context.Background(), restaurantID); e == nil {
					restaurantName = name
				}
			}
			mailItems := make([]mailer.OrderItem, len(items))
			for i, it := range items {
				mailItems[i] = mailer.OrderItem{Name: it.Name, Quantity: it.Quantity}
			}
			_ = s.mailer.SendOrderConfirmation(email, restaurantName, tableNumber, mailItems, notes, confirmURL)
		}()
	}

	return toResponse(o), nil
}

func (s *Service) ConfirmOrder(ctx context.Context, token string) (*ConfirmResponse, error) {
	conf, err := s.repo.FindConfirmationByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("find confirmation: %w", err)
	}
	if conf == nil || conf.ExpiresAt.Before(time.Now()) {
		return nil, ErrNotFound
	}

	if err := s.repo.DeleteConfirmationByToken(ctx, token); err != nil {
		return nil, fmt.Errorf("delete confirmation: %w", err)
	}

	o, err := s.repo.FindByID(ctx, conf.OrderID)
	if err != nil {
		return nil, fmt.Errorf("find order: %w", err)
	}
	if o == nil {
		return nil, ErrNotFound
	}

	if err := s.repo.UpdateStatus(ctx, o.ID, StatusPending); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update status: %w", err)
	}
	o.Status = StatusPending

	resp := toResponse(o)
	// Broadcast to restaurant (staff panel sees new order) and table channel (guest sees status change)
	s.hub.BroadcastToRestaurantAndTable(o.RestaurantID, o.TableID, ws.Event{
		Type:    ws.EventOrderStatusChanged,
		Payload: resp,
	})

	return &ConfirmResponse{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		TableID:      o.TableID,
	}, nil
}

func (s *Service) GetActiveByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	orders, err := s.repo.FindActiveByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(orders))
	for i, o := range orders {
		responses[i] = *toResponse(&o)
	}
	return responses, nil
}

func (s *Service) GetByRestaurant(ctx context.Context, restaurantID string) ([]Response, error) {
	orders, err := s.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	var responses []Response
	for _, o := range orders {
		if o.Status != StatusAwaitingConfirmation {
			responses = append(responses, *toResponse(&o))
		}
	}
	return responses, nil
}

func (s *Service) GetByTable(ctx context.Context, restaurantID, tableID string) ([]Response, error) {
	orders, err := s.repo.FindByTableID(ctx, restaurantID, tableID)
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

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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
