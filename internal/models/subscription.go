package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ServiceName string     `gorm:"not null;index" json:"service_name" binding:"required"`
	Price       int        `gorm:"not null" json:"price" binding:"required,gte=0"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" binding:"required"`
	StartDate   time.Time  `gorm:"not null;type:date" json:"start_date" binding:"required"`
	EndDate     *time.Time `gorm:"type:date" json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required"`
	Price       int     `json:"price" binding:"required,gte=0"`
	UserID      string  `json:"user_id" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     *string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type SubscriptionFilter struct {
	UserID      string `form:"user_id"`
	ServiceName string `form:"service_name"`
}

type PeriodRequest struct {
	StartDate   string `form:"start_date" binding:"required"`
	EndDate     string `form:"end_date" binding:"required"`
	UserID      string `form:"user_id"`
	ServiceName string `form:"service_name"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}
