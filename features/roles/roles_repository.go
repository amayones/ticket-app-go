package roles

import (
	"context"
	"database/sql"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
)

// RepositoryInterface reads/writes CPROLE + permission mapping.
type RepositoryInterface interface {
	List(ctx context.Context) ([]models.Role, error)
	GetByCode(ctx context.Context, code string) (*models.Role, error)
	RoleExists(ctx context.Context, code string) (bool, error)
	Create(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, code string) error
	Count(ctx context.Context) (int, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	GetRolePermissions(ctx context.Context, roleCode string) ([]string, error)
	SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error
	HasPermission(ctx context.Context, roleCode, permCode string) (bool, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) roleTable() string { return r.dialect.Table("CPROLE") }
func (r *Repository) permTable() string { return r.dialect.Table("CPPERMISSION") }
func (r *Repository) mapTable() string  { return r.dialect.Table("CPROLEPERMISSION") }

func (r *Repository) List(ctx context.Context) ([]models.Role, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM `+r.roleTable()+` ORDER BY CODE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]models.Role, 0)
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) RoleExists(ctx context.Context, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.roleTable()+` WHERE CODE = ?`), code).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var role models.Role
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM `+r.roleTable()+` WHERE CODE = ?`), code,
	).Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) Create(ctx context.Context, role *models.Role) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`INSERT INTO `+r.roleTable()+` (CODE, NAME) VALUES (?, ?)`), role.Code, role.Name)
	return err
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		r.dialect.Bind(`DELETE FROM `+r.roleTable()+` WHERE CODE = ?`), code)
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

func (r *Repository) Count(ctx context.Context) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+r.roleTable()).Scan(&n)
	return n, err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT ID, CODE, NAME, PERMGROUP, DESCRIPTION, CREATED_AT FROM `+r.permTable()+` ORDER BY PERMGROUP ASC, CODE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perms := make([]models.Permission, 0)
	for rows.Next() {
		var p models.Permission
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Group, &desc, &p.CreatedAt); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(p.Code, "MENU_") {
			continue
		}
		p.Description = desc.String
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(
		`SELECT PERMISSION_CODE FROM `+r.mapTable()+` WHERE ROLE_CODE = ? ORDER BY PERMISSION_CODE ASC`), roleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		if strings.HasPrefix(c, "MENU_") {
			codes = append(codes, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return codes, nil
}

// SetRolePermissions replaces the permission set atomically.
func (r *Repository) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.mapTable()+` WHERE ROLE_CODE = ?`), roleCode); err != nil {
		return err
	}
	for _, pc := range permCodes {
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(
			`INSERT INTO `+r.mapTable()+` (ROLE_CODE, PERMISSION_CODE) VALUES (?, ?)`),
			roleCode, pc); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) HasPermission(ctx context.Context, roleCode, permCode string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.mapTable()+` WHERE ROLE_CODE = ? AND PERMISSION_CODE = ?`),
		roleCode, permCode).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
