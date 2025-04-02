package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Eduardo-Barreto/web-ponderada/backend/models"
)

type postgresProductRepository struct {
	db *pgxpool.Pool
}

// NewPostgresProductRepository creates a new instance of ProductRepository
func NewPostgresProductRepository(db *pgxpool.Pool) ProductRepository {
	return &postgresProductRepository{db: db}
}

func (r *postgresProductRepository) CreateProduct(ctx context.Context, product *models.Product) (int, error) {
	query := `INSERT INTO products (description, value, quantity, image, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	now := time.Now()
	err := r.db.QueryRow(ctx, query, product.Description, product.Value, product.Quantity, product.Image, now, now).Scan(&product.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}
	return product.ID, nil
}

func (r *postgresProductRepository) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	query := `SELECT id, description, value, quantity, image, created_at, updated_at
	          FROM products WHERE id = $1`
	product := &models.Product{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.Description, &product.Value, &product.Quantity, &product.Image, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return product, nil
}

func (r *postgresProductRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	query := `SELECT id, description, value, quantity, image, created_at, updated_at FROM products ORDER BY description ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Description, &p.Value, &p.Quantity, &p.Image, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}

func (r *postgresProductRepository) UpdateProduct(ctx context.Context, id int, productInput *models.ProductInput, imageFilename *string) error {
	query := `UPDATE products SET description = $1, value = $2, quantity = $3, updated_at = $4`
	args := []interface{}{productInput.Description, productInput.Value, productInput.Quantity, time.Now()}
	argID := 5 // Start arg index after fixed fields

	if imageFilename != nil {
		query += fmt.Sprintf(", image = $%d", argID)
		args = append(args, *imageFilename)
		argID++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argID)
	args = append(args, id)

	cmdTag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("product not found or no changes made")
	}
	return nil
}

func (r *postgresProductRepository) DeleteProduct(ctx context.Context, id int) error {
	// Important: We might need to delete the associated image file first.
	// This logic should ideally be in a service layer, not directly in the repo.
	// For now, we just delete the DB record.
	query := `DELETE FROM products WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("product not found")
	}
	return nil
}
