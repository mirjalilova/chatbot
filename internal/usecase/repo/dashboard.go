package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/pkg/postgres"
)

type DashboardRepo struct {
	pg     *postgres.Postgres
	config *config.Config
}

func NewDashboardRepo(pg *postgres.Postgres, config *config.Config) *DashboardRepo {
	return &DashboardRepo{
		pg:     pg,
		config: config,
	}
}

func (r *DashboardRepo) GetUserAndRequestCount(ctx context.Context, fromDate, toDate time.Time) (*[]entity.DashboardActiveUsers, error) {
	query := `
		SELECT * FROM get_daily_chat_stats($1, $2)
	`

	var stats []entity.DashboardActiveUsers
	rows, err := r.pg.Pool.Query(ctx, query, fromDate, toDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no data found")
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat entity.DashboardActiveUsers
		var date time.Time
		if err := rows.Scan(&date, &stat.ActiveUsers, &stat.RequestsCount); err != nil {
			return nil, err
		}
		stat.Day = date.Format("2006-01-02")
		stats = append(stats, stat)
	}

	return &stats, nil
}
