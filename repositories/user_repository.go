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
// Identity is the public CODE (CPUSER.CODE); numeric IDs never leave the DB.
type UserRepositoryInterface interface {
	List(ctx context.Context, limit, offset int) ([]models.User, error)
	GetByCode(ctx context.Context, code string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, code string) error
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

const userColumns = `u.CODE, u.USERNAME, u.EMAIL, u.PASSWORD, u.ROLE_CODE, r.NAME, u.CREATED_AT, u.UPDATED_AT`
const userJoin = `FROM dbo.CPUSER u JOIN dbo.CPROLE r ON r.CODE = u.ROLE_CODE`

func scanUser(row interface {
	Scan(dest ...any) error
}, user *models.User) error {
	return row.Scan(
		&user.Code,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RoleCode,
		&user.RoleName,
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
		SELECT ` + userColumns + `
		` + userJoin + `
		ORDER BY u.ID DESC
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

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
		VALUES (@p1, @p2, @p3, @p4, @p5)
	`
	_, err := r.db.ExecContext(ctx, query, user.Code, user.Username, user.Email, user.Password, user.RoleCode)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		UPDATE dbo.CPUSER
		SET USERNAME = @p1, EMAIL = @p2, PASSWORD = @p3, UPDATED_AT = GETDATE()
		WHERE CODE = @p4
	`
	res, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.Code)
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

func (r *UserRepository) Delete(ctx context.Context, code string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPUSER WHERE CODE = @p1`, code)
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

func (r *UserRepository) GetByCode(ctx context.Context, code string) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT ` + userColumns + `
		` + userJoin + `
		WHERE u.CODE = @p1
	`
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, code), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT ` + userColumns + `
		` + userJoin + `
		WHERE u.USERNAME = @p1
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
		SELECT ` + userColumns + `
		` + userJoin + `
		WHERE u.EMAIL = @p1
	`
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, email), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
