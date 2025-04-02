package repository

import (
	"context"
	"mime/multipart"

	"github.com/Eduardo-Barreto/web-ponderada/backend/models"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, id int, updateData *models.UserUpdateInput) error
	UpdateUserProfilePic(ctx context.Context, id int, filename string) error
	DeleteUser(ctx context.Context, id int) error
}

// ProductRepository defines methods for product data access
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) (int, error)
	GetProductByID(ctx context.Context, id int) (*models.Product, error)
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	UpdateProduct(ctx context.Context, id int, product *models.ProductInput, imageFilename *string) error
	DeleteProduct(ctx context.Context, id int) error
}

// StorageRepository defines methods for file storage (could be local, S3, etc.)
type StorageRepository interface {
	SaveFile(ctx context.Context, file *multipart.FileHeader, destination string) (string, error) // returns generated filename
	GetFilePath(filename string) string
	DeleteFile(ctx context.Context, filename string) error
}

