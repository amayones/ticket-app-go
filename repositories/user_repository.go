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
	UpdateRole(ctx context.Context, code, roleCode string) error
	Delete(ctx context.Context, code string) error
	Count(ctx context.Context) (int, error)
	CountByRole(ctx context.Context, roleCode string) (int, error)
}

type UserRepository struct {
	db      *sql.DB
	dialect Dialect
}

func NewUserRepository(db *sql.DB, dialect Dialect) UserRepositoryInterface {
	return &UserRepository{db: db, dialect: dialect}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}

const userColumns = `u.CODE, u.USERNAME, u.EMAIL, u.PASSWORD, u.ROLE_CODE, r.NAME, u.CREATED_AT, u.UPDATED_AT`

func (r *UserRepository) userFrom() string {
	return `FROM ` + r.dialect.Table("CPUSER") + ` u JOIN ` +
		r.dialect.Table("CPROLE") + ` r ON r.CODE = u.ROLE_CODE`
}

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
	query := `SELECT ` + userColumns + ` ` + r.userFrom() + ` ORDER BY u.ID DESC `
	var args []any
	if r.dialect == DialectMSSQL {
		query += pageMSSQL()
		args = []any{offset, limit}
	} else {
		query += pageStd()
		args = []any{limit, offset}
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
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
	query := `INSERT INTO ` + r.dialect.Table("CPUSER") + `
		(CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
		VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		user.Code, user.Username, user.Email, user.Password, user.RoleCode)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.dialect.Table("CPUSER") + `
		SET USERNAME = ?, EMAIL = ?, PASSWORD = ?, UPDATED_AT = ` + r.dialect.Now() + `
		WHERE CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		user.Username, user.Email, user.Password, user.Code)
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

func (r *UserRepository) UpdateRole(ctx context.Context, code, roleCode string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.dialect.Table("CPUSER") + ` SET ROLE_CODE = ?, UPDATED_AT = ` +
		r.dialect.Now() + ` WHERE CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), roleCode, code)
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
	query := `DELETE FROM ` + r.dialect.Table("CPUSER") + ` WHERE CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), code)
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

func (r *UserRepository) userBy(ctx context.Context, where string, arg any) (*models.User, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `SELECT ` + userColumns + ` ` + r.userFrom() + ` WHERE ` + where
	var user models.User
	if err := scanUser(r.db.QueryRowContext(ctx, r.dialect.Bind(query), arg), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByCode(ctx context.Context, code string) (*models.User, error) {
	return r.userBy(ctx, `u.CODE = ?`, code)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.userBy(ctx, `u.USERNAME = ?`, username)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.userBy(ctx, `u.EMAIL = ?`, email)
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+r.dialect.Table("CPUSER")).Scan(&n)
	return n, err
}

func (r *UserRepository) CountByRole(ctx context.Context, roleCode string) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.dialect.Table("CPUSER")+` WHERE ROLE_CODE = ?`), roleCode).Scan(&n)
	return n, err
}
