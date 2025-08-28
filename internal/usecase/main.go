package usecase

import (
	"chatbot/config"
	"chatbot/internal/usecase/repo"
	"chatbot/pkg/postgres"
)

type UseCase struct {
	UserRepo        UserRepoI
	RestrictionRepo RestrictionRepoI
	ChatRepo        ChatRepoI
	PDFRepo         PDFRepoI
	DashboardRepo   DashboardRepoI
}

func New(pg *postgres.Postgres, config *config.Config) *UseCase {
	return &UseCase{
		UserRepo:        repo.NewUserRepo(pg, config),
		RestrictionRepo: repo.NewRestrictionRepo(pg, config),
		ChatRepo:        repo.NewChatRepo(pg, config),
		DashboardRepo:   repo.NewDashboardRepo(pg, config),
	}
}
