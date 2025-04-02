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

type postgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new instance of UserRepository
func NewPostgresUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) (int, error) {
	query := `INSERT INTO users (name, email, password_hash, profile_pic, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	now := time.Now()
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.Password, user.ProfilePic, now, now).Scan(&user.ID)
	if err != nil {
		// TODO: Check for unique constraint violation on email
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return user.ID, nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, profile_pic, created_at, updated_at
	          FROM users WHERE email = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.ProfilePic, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found is not necessarily an error here
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, profile_pic, created_at, updated_at
	          FROM users WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.ProfilePic, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found") // More explicit error for GetByID
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, name, email, profile_pic, created_at, updated_at FROM users ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.ProfilePic, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, id int, updateData *models.UserUpdateInput) error {
	// Build the update query dynamically based on provided fields
	query := "UPDATE users SET updated_at = $1"
	args := []interface{}{time.Now()}
	argID := 2 // Start argument index after updated_at

	if updateData.Name != nil {
		query += fmt.Sprintf(", name = $%d", argID)
		args = append(args, *updateData.Name)
		argID++
	}
	if updateData.Email != nil {
		query += fmt.Sprintf(", email = $%d", argID)
		args = append(args, *updateData.Email)
		argID++
	}

	// Only proceed if there's something to update
	if len(args) == 1 {
		return errors.New("no update fields provided")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argID)
	args = append(args, id)

	cmdTag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		// TODO: Check for unique constraint violation on email if updated
		return fmt.Errorf("failed to update user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found or no changes made")
	}

	return nil
}

func (r *postgresUserRepository) UpdateUserProfilePic(ctx context.Context, id int, filename string) error {
	query := `UPDATE users SET profile_pic = $1, updated_at = $2 WHERE id = $3`
	cmdTag, err := r.db.Exec(ctx, query, filename, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update profile picture: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}


func (r *postgresUserRepository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
