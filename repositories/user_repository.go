package repositories

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
)

const (
	DefaultListLimit = 50
	MaxListLimit     = 200
	queryTimeout     = 5 * time.Second
)

// UserRepositoryInterface is the contract services depend on (mockable).
type UserRepositoryInterface interface {
	List(ctx context.Context, limit, offset int) ([]models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (int, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}

func scanUser(row interface {
	Scan(dest ...any) error
}, user *models.User) error {
	return row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

// List returns users without password hashes, paginated (DoS-safe).
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, username, email, '' AS password, created_at, updated_at
		FROM users
		ORDER BY id DESC
		OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY
	`
	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := scanUser(rows, &user); err != nil {
			return nil, err
		}
		user.Password = ""
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO users (username, email, password)
		OUTPUT INSERTED.id
		VALUES (@p1, @p2, @p3)
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}
	user.ID = id
	return id, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		UPDATE users
		SET username = @p1, email = @p2, password = @p3, updated_at = GETDATE()
		WHERE id = @p4
	`
	res, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = @p1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = @p1
	`
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, id), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = @p1
	`
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, username), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE email = @p1
	`
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, email), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
