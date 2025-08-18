package models

import (
	"time"
)

type Budget struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	CategoryID uint      `json:"category_id" gorm:"not null"`
	Month      uint      `json:"month" gorm:"not null"`
	Year       uint      `json:"year" gorm:"not null"`
	Amount     uint      `json:"amount" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Budget) TableName() string {
	return "budgets"
}

type PublicBudget struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	CategoryID uint      `json:"category_id"`
	Month      uint      `json:"month"`
	Year       uint      `json:"year"`
	Amount     uint      `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Budget) ToPublicBudget() PublicBudget {
	return PublicBudget{
		ID:         c.ID,
		UserID:     c.UserID,
		CategoryID: c.CategoryID,
		Month:      c.Month,
		Year:       c.Year,
		Amount:     c.Amount,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}
