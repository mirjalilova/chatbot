package handler

import (
	"chatbot/config"
	"chatbot/internal/usecase"

	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	Config       *config.Config
	UseCase      *usecase.UseCase
	GeminiClient *genai.Client
	Redis        *redis.Client
}

func NewHandler(c *config.Config, useCase *usecase.UseCase, geminiClient *genai.Client, rdb *redis.Client) *Handler {
	return &Handler{
		Config:       c,
		UseCase:      useCase,
		GeminiClient: geminiClient,
		Redis:        rdb,
	}
}
