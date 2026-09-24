package repositories

import (
	"context"
	"database/sql"

	"golang-backend/models"
)

// RoleRepositoryInterface reads/writes CPROLE + permission mapping.
type RoleRepositoryInterface interface {
	List(ctx context.Context) ([]models.Role, error)
	GetByCode(ctx context.Context, code string) (*models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, code string) error
	Count(ctx context.Context) (int, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	GetRolePermissions(ctx context.Context, roleCode string) ([]string, error)
	SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error
	HasPermission(ctx context.Context, roleCode, permCode string) (bool, error)
}

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepositoryInterface {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) List(ctx context.Context) ([]models.Role, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM dbo.CPROLE ORDER BY CODE ASC`)
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

func (r *RoleRepository) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var role models.Role
	err := r.db.QueryRowContext(ctx,
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM dbo.CPROLE WHERE CODE = @p1`, code,
	).Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (@p1, @p2)`, role.Code, role.Name)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, code string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPROLE WHERE CODE = @p1`, code)
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

func (r *RoleRepository) Count(ctx context.Context) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.CPROLE`).Scan(&n)
	return n, err
}

func (r *RoleRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT ID, CODE, NAME, PERMGROUP, DESCRIPTION, CREATED_AT FROM dbo.CPPERMISSION ORDER BY PERMGROUP ASC, CODE ASC`)
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
		p.Description = desc.String
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT PERMISSION_CODE FROM dbo.CPROLEPERMISSION WHERE ROLE_CODE = @p1 ORDER BY PERMISSION_CODE ASC`, roleCode)
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
		codes = append(codes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return codes, nil
}

// SetRolePermissions replaces the permission set atomically.
func (r *RoleRepository) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM dbo.CPROLEPERMISSION WHERE ROLE_CODE = @p1`, roleCode); err != nil {
		return err
	}
	for _, pc := range permCodes {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO dbo.CPROLEPERMISSION (ROLE_CODE, PERMISSION_CODE) VALUES (@p1, @p2)`, roleCode, pc); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RoleRepository) HasPermission(ctx context.Context, roleCode, permCode string) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM dbo.CPROLEPERMISSION WHERE ROLE_CODE = @p1 AND PERMISSION_CODE = @p2`,
		roleCode, permCode).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
