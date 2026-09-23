package repositories

import (
	"database/sql"

	"golang-backend/models"
)

type UserRepositoryInterface interface {
	GetAll() ([]models.User, error)
	GetByID(id int) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id int) error
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		ORDER BY id DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES (@p1, @p2, @p3)
	`
	_, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
	)
	return err
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = @p1, email = @p2, password = @p3, updated_at = GETDATE()
		WHERE id = @p4
	`
	res, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.ID,
	)
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

func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = @p1`
	res, err := r.DB.Exec(query, id)
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

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = @p1
	`
	var user models.User
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = @p1
	`
	var user models.User
	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
