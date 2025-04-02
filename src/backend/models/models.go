package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"-"` // Never expose password hash
	ProfilePic string    `json:"profile_pic"` // Stores filename or path/URL
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input struct for user registration (doesn't include hashed password)
type UserRegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"` // Add validation
}

// Input struct for user update
type UserUpdateInput struct {
	Name *string `json:"name"` // Use pointers to distinguish between empty and not provided
	Email *string `json:"email" binding:"omitempty,email"` // Optional email update
}

// Input struct for login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Product struct {
	ID          int       `json:"id"`
	Description string    `json:"description" binding:"required"`
	Value       float64   `json:"value" binding:"required,gt=0"` // Must be greater than 0
	Quantity    int       `json:"quantity" binding:"required,gte=0"` // Must be 0 or more
	Image       string    `json:"image"` // Stores filename or path/URL
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Input struct for product creation/update (excluding ID and timestamps)
type ProductInput struct {
	Description string  `json:"description" binding:"required"`
	Value       float64 `json:"value" binding:"required,gt=0"`
	Quantity    int     `json:"quantity" binding:"required,gte=0"`
	// Image is handled separately via multipart form
}
