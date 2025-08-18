package models

import (
	"time"
)

type Transaction struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	CategoryID uint      `json:"category_id" gorm:"not null"`
	Amount     uint      `json:"amount" gorm:"not null"`
	Type       string    `json:"type" gorm:"not null"`
	Remarks    string    `json:"remarks" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type PublicTransaction struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	CategoryID uint      `json:"category_id"`
	Amount     uint      `json:"amount"`
	Type       string    `json:"type"`
	Remarks    string    `json:"remarks"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Transaction) ToPublicTransaction() PublicTransaction {
	return PublicTransaction{
		ID:         c.ID,
		UserID:     c.UserID,
		CategoryID: c.CategoryID,
		Amount:     c.Amount,
		Type:       c.Type,
		Remarks:    c.Remarks,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}
