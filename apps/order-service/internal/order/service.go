package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type ProductSnapshot struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Active bool    `json:"active"`
}
type CreateItem struct {
	ProductID string `json:"productId" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
type CreateInput struct {
	CustomerID string       `json:"customerId" validate:"required,uuid"`
	Items      []CreateItem `json:"items" validate:"required,min=1,dive"`
}

var ErrNotFound = errors.New("order not found")

type Service struct {
	db         *gorm.DB
	catalogURL string
	client     *http.Client
}

func NewService(db *gorm.DB, url string) *Service {
	return &Service{db, url, &http.Client{Timeout: 5e9}}
}
func (s *Service) Create(ctx context.Context, in CreateInput) (*Order, error) {
	o := &Order{ID: uuid.NewString(), CustomerID: in.CustomerID, Status: "CREATED"}
	for _, requested := range in.Items {
		p, err := s.getProduct(ctx, requested.ProductID)
		if err != nil {
			return nil, err
		}
		item := Item{ID: uuid.NewString(), OrderID: o.ID, ProductID: p.ID, ProductName: p.Name, Quantity: requested.Quantity, UnitPrice: p.Price, Subtotal: p.Price * float64(requested.Quantity)}
		o.Total += item.Subtotal
		o.Items = append(o.Items, item)
	}
	payload, _ := json.Marshal(o)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(o).Error; err != nil {
			return err
		}
		if err := tx.Create(&o.Items).Error; err != nil {
			return err
		}
		return tx.Create(&OutboxEvent{AggregateType: "order", AggregateID: o.ID, EventType: "order.created", Payload: payload}).Error
	})
	return o, err
}
func (s *Service) List(ctx context.Context) ([]Order, error) {
	var items []Order
	err := s.db.WithContext(ctx).Preload("Items").Order("created_at desc").Find(&items).Error
	return items, err
}
func (s *Service) Get(ctx context.Context, id string) (*Order, error) {
	var o Order
	err := s.db.WithContext(ctx).Preload("Items").First(&o, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &o, err
}
func (s *Service) getProduct(ctx context.Context, id string) (ProductSnapshot, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/products/%s", s.catalogURL, id), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return ProductSnapshot{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ProductSnapshot{}, fmt.Errorf("catalog returned %d", resp.StatusCode)
	}
	var p ProductSnapshot
	err = json.NewDecoder(resp.Body).Decode(&p)
	if !p.Active {
		return ProductSnapshot{}, errors.New("inactive product")
	}
	return p, err
}
