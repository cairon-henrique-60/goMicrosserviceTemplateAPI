package product

import "time"

type Product struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SKU         string    `gorm:"size:80;uniqueIndex;not null" json:"sku"`
	Name        string    `gorm:"size:180;not null" json:"name"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Price       float64   `gorm:"type:numeric(14,2);not null" json:"price"`
	Stock       int       `gorm:"not null" json:"stock"`
	Active      bool      `gorm:"not null" json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type OutboxEvent struct {
	ID            string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AggregateType string
	AggregateID   string `gorm:"type:uuid"`
	EventType     string
	Payload       []byte `gorm:"type:jsonb"`
	CreatedAt     time.Time
	PublishedAt   *time.Time
}
