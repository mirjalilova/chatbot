package handler

import (
	"chatbot/config"
	"chatbot/internal/usecase"

	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
	"chatbot/pkg/minio"
)

type Handler struct {
	Config       *config.Config
	UseCase      *usecase.UseCase
	GeminiClient *genai.Client
	Redis        *redis.Client
	MinIO        *minio.MinIO
}

func NewHandler(c *config.Config, useCase *usecase.UseCase, geminiClient *genai.Client, rdb *redis.Client, mn minio.MinIO) *Handler {
	return &Handler{
		Config:       c,
		UseCase:      useCase,
		GeminiClient: geminiClient,
		Redis:        rdb,
		MinIO:        &mn,
	}
}
