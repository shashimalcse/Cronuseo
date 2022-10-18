package models

type Resource struct {
	ID          int    `json:"id" gorm:"primary_key"`
	Key         string `json:"key" binding:"required,min=4"`
	Name        string `json:"name" binding:"required,min=4"`
	Description string `json:"description"`
	ProjectID   int    `json:"-" gorm:"foreignKey:ID"`
}