package models

import (
	"time"

	"ashborn.id/moniplan/database"
)

type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null;size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryAndBudget struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CategoryID uint      `json:"category_id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	Name       string    `json:"name" gorm:"not null;size:100"`
	Month      uint      `json:"month" gorm:"not null"`
	Year       uint      `json:"year" gorm:"not null"`
	Amount     uint      `json:"amount" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

type PublicCategory struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Category) ToPublicCategory() PublicCategory {
	return PublicCategory{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (c *CategoryAndBudget) ToPublicCategoryAndBudget() CategoryAndBudget {
	return CategoryAndBudget{
		CategoryID: c.CategoryID,
		UserID:     c.UserID,
		Name:       c.Name,
		Month:      c.Month,
		Year:       c.Year,
		Amount:     c.Amount,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func (c *CategoryAndBudget) GetCategoryAndBudgetByID(categoryID uint) CategoryAndBudget {
	var category Category
	var budget Budget

	if errCategory := database.DB.First(&category, categoryID).Error; errCategory != nil {
		category = Category{}
	}

	if errBudget := database.DB.Where("category_id = ?", categoryID).Last(&budget).Error; errBudget != nil {
		budget = Budget{}
	}

	return CategoryAndBudget{
		CategoryID: categoryID,
		UserID:     category.UserID,
		Name:       category.Name,
		Month:      budget.Month,
		Year:       budget.Month,
		Amount:     budget.Month,
		CreatedAt:  category.CreatedAt,
		UpdatedAt:  category.UpdatedAt,
	}
}
