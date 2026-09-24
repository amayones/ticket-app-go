package repositories

import (
	"context"
	"database/sql"

	"golang-backend/models"
)

// RoleRepositoryInterface reads the CPROLE master table.
type RoleRepositoryInterface interface {
	List(ctx context.Context) ([]models.Role, error)
	GetByCode(ctx context.Context, code string) (*models.Role, error)
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
