package repo

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/pkg/postgres"
)

type RestrictionRepo struct {
	pg     *postgres.Postgres
	config *config.Config
}

func NewRestrictionRepo(pg *postgres.Postgres, config *config.Config) *RestrictionRepo {
	return &RestrictionRepo{
		pg:     pg,
		config: config,
	}
}

func (r *RestrictionRepo) GetById(ctx context.Context, id *entity.ById) (*entity.Restriction, error) {
	query := `
		SELECT id, type, request_limit
		FROM restrictions
		WHERE id = $1`

	var res entity.Restriction
	err := r.pg.Pool.QueryRow(ctx, query, id.Id).Scan(
		&res.ID,
		&res.Type,
		&res.RequestLimit,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("restriction not found")
		}
		return nil, err
	}

	return &res, nil
}

func (r *RestrictionRepo) GetAll(ctx context.Context, filter *entity.Filter) (*entity.ListRestriction, error) {
	query := `
		SELECT id, type, request_limit
		FROM restrictions
		ORDER BY type`

	if filter.Limit > 0 {
		query += " LIMIT $1 OFFSET $2"
	}

	var args []interface{}

	if filter.Limit != 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, filter.Limit)
		args = append(args, filter.Offset)
	}

	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user list is empty")
		}
		return nil, err
	}
	defer rows.Close()

	var result entity.ListRestriction
	for rows.Next() {
		var r entity.Restriction
		err := rows.Scan(&r.ID, &r.Type, &r.RequestLimit)
		if err != nil {
			return nil, err
		}
		result.Restrictions = append(result.Restrictions, r)
	}

	return &result, nil
}

func (r *RestrictionRepo) Update(ctx context.Context, req *entity.UpdateRestriction) error {
	query := `UPDATE restrictions SET`
	var args []interface{}
	var sets []string

	if req.RequestLimit != nil {
		sets = append(sets, " request_limit = $"+strconv.Itoa(len(args)+1))
		args = append(args, *req.RequestLimit)
	}

	// if req.TimeLimit != nil {
	// 	sets = append(sets, " time_limit = $"+strconv.Itoa(len(args)+1))
	// 	args = append(args, *req.TimeLimit)
	// }

	if len(sets) == 0 {
		return errors.New("no fields to update")
	}

	query += strings.Join(sets, ", ") + " WHERE id = $" + strconv.Itoa(len(args)+1)
	args = append(args, req.ID)

	_, err := r.pg.Pool.Exec(ctx, query, args...)
	return err
}

// func (r *RestrictionRepo) Delete(ctx context.Context, id *entity.ById) error {
// 	_, err := r.pg.Pool.Exec(ctx, `UPDATE users SET deleted_at = EXTRACT(EPOCH FROM NOW()) WHERE id = $1`, id.Id)
// 	return err
// }
