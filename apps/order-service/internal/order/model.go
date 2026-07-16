package order

import "time"

type Order struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerID string    `gorm:"type:uuid;not null" json:"customerId"`
	Status     string    `gorm:"size:40;not null" json:"status"`
	Total      float64   `gorm:"type:numeric(14,2);not null" json:"total"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
type Item struct {
	ID          string  `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID     string  `gorm:"type:uuid;index;not null" json:"-"`
	ProductID   string  `gorm:"type:uuid;not null" json:"productId"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Subtotal    float64 `json:"subtotal"`
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
