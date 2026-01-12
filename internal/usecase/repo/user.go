package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/pkg/postgres"
)

type UserRepo struct {
	pg     *postgres.Postgres
	config *config.Config
}

// New -.
func NewUserRepo(pg *postgres.Postgres, config *config.Config) *UserRepo {
	return &UserRepo{
		pg:     pg,
		config: config,
	}
}

func (r *UserRepo) Create(ctx context.Context, req *entity.CreateUser) (*entity.UserInfo, error) {
	var id string

	query := `
		INSERT INTO users (
			phone_number
		) VALUES($1)
		RETURNING id`

	err := r.pg.Pool.QueryRow(ctx, query, req.PhoneNumber).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &entity.UserInfo{
		ID: id,
	}, nil
}

func (r *UserRepo) CreateGuest(ctx context.Context) (string, error) {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (role)
		VALUES ('guest')
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create guest: %w", err)
	}

	var chatRoomID string
	err = tx.QueryRow(ctx, `
		INSERT INTO chat_rooms (user_id)
		VALUES ($1)
		RETURNING id
	`, userID).Scan(&chatRoomID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat room: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit tx: %w", err)
	}

	return userID, nil
}


// func (r *UserRepo) Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginRes, error) {
// 	query := `
// 		SELECT
// 			password,
// 			id
// 		FROM
// 			users
// 		WHERE
// 			phone_number = $1
// 		AND
// 			deleted_at = 0
// 	`
// 	row := r.pg.Pool.QueryRow(ctx, query, req.Login)
// 	var password string
// 	var id string
// 	err := row.Scan(&password, &id)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, errors.New("user not found")
// 		}
// 		return nil, err
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
// 		return nil, errors.New("invalid login or password")
// 	}

// 	token := token.GenerateJWTToken(id)

// 	return &entity.LoginRes{Token: token.AccessToken,
// 		Message: "success"}, nil
// }

func (r *UserRepo) GetById(ctx context.Context, req *entity.ById) (*entity.UserInfo, error) {

	var res entity.UserInfo
	var createdAt time.Time

	query := `
	SELECT
		id,
		full_name,
		phone_number,
		avatar,
		role,
		created_at
	FROM 
		users
	WHERE 
		deleted_at = 0
	AND 
		id = $1
	`

	row := r.pg.Pool.QueryRow(ctx, query, req.Id)
	err := row.Scan(
		&res.ID,
		&res.FullName,
		&res.PhoneNumber,
		&res.Avatar,
		&res.Role,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	res.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

	return &res, nil
}

func (r *UserRepo) GetAll(ctx context.Context, req *entity.Filter, status string) (*entity.UserList, error) {

	resp := &entity.UserList{}

	query := `
	SELECT
		COUNT(id) OVER () AS total_count,
		id,
		full_name,
		phone_number,
		created_at
	FROM
		users
	WHERE
		deleted_at = 0
	`

	var args []interface{}

	if status != "" {
		query += " AND status = $" + strconv.Itoa(len(args)+1)
		args = append(args, status)
	}

	if req.Limit != 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, req.Limit)
		args = append(args, req.Offset)
	}

	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user list is empty")
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		res := entity.UserInfo{}
		var count int
		var createdAt time.Time

		err := rows.Scan(
			&count,
			&res.ID,
			&res.FullName,
			&res.PhoneNumber,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}
		res.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		resp.Users = append(resp.Users, res)
		resp.Count = count
	}

	return resp, nil
}

func (r *UserRepo) Update(ctx context.Context, req *entity.UpdateUser) error {
	query := `
	UPDATE
		users
	SET`

	var conditions []string
	var args []interface{}

	if req.FullName != "" && req.FullName != "string" {
		conditions = append(conditions, " full_name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.FullName)
	}
	if req.PhoneNumber != "" && req.PhoneNumber != "string" {
		conditions = append(conditions, " phone_number = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.PhoneNumber)
	}
	if req.Avatar != "" && req.Avatar != "string" {
		conditions = append(conditions, " avatar = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Avatar)
	}

	conditions = append(conditions, " updated_at = CURRENT_TIMESTAMP")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	_, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, req *entity.ById) error {

	_, err := r.pg.Pool.Exec(ctx, `UPDATE users SET deleted_at = EXTRACT(EPOCH FROM NOW()) WHERE id = $1`, req.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) CheckExist(ctx context.Context, phone string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone_number = $1 AND deleted_at = 0)`
	err := r.pg.Pool.QueryRow(ctx, query, phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*entity.GetByPhone, error) {
	var res entity.GetByPhone
	query := `
		SELECT
			id,
			role
		FROM
			users
		WHERE
			phone_number = $1
		AND
			deleted_at = 0
	`
	row := r.pg.Pool.QueryRow(ctx, query, phone)
	err := row.Scan(
		&res.Id,
		&res.Role,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *UserRepo) GetMe(ctx context.Context, id string) (*entity.GetMe, error) {
	var res entity.GetMe
	query := `
		SELECT
			full_name,
			role,
			avatar,
			language
		FROM
			users
		WHERE
			id = $1
		AND
			deleted_at = 0
	`
	row := r.pg.Pool.QueryRow(ctx, query, id)
	err := row.Scan(
		&res.FullName,
		&res.Role,
		&res.Avatar,
		&res.Language,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}