package product

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

var ErrNotFound = errors.New("product not found")

type CreateInput struct {
	SKU         string  `json:"sku" validate:"required,min=2,max=80"`
	Name        string  `json:"name" validate:"required,min=2,max=180"`
	Description string  `json:"description" validate:"max=2000"`
	Price       float64 `json:"price" validate:"gte=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
}
type UpdateInput struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Active      *bool    `json:"active"`
}
type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db} }
func (s *Service) Create(ctx context.Context, in CreateInput) (*Product, error) {
	p := &Product{ID: uuid.NewString(), SKU: strings.TrimSpace(in.SKU), Name: strings.TrimSpace(in.Name), Description: strings.TrimSpace(in.Description), Price: in.Price, Stock: in.Stock, Active: true}
	payload, _ := json.Marshal(p)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		return tx.Create(&OutboxEvent{AggregateType: "product", AggregateID: p.ID, EventType: "product.created", Payload: payload}).Error
	})
	return p, err
}
func (s *Service) List(ctx context.Context) ([]Product, error) {
	var items []Product
	err := s.db.WithContext(ctx).Order("created_at desc").Find(&items).Error
	return items, err
}
func (s *Service) Get(ctx context.Context, id string) (*Product, error) {
	var p Product
	err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*Product, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		p.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		p.Description = strings.TrimSpace(*in.Description)
	}
	if in.Price != nil {
		p.Price = *in.Price
	}
	if in.Stock != nil {
		p.Stock = *in.Stock
	}
	if in.Active != nil {
		p.Active = *in.Active
	}
	payload, _ := json.Marshal(p)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(p).Error; err != nil {
			return err
		}
		return tx.Create(&OutboxEvent{AggregateType: "product", AggregateID: p.ID, EventType: "product.updated", Payload: payload}).Error
	})
	return p, err
}
func (s *Service) Delete(ctx context.Context, id string) error {
	p, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(p).Error
}
