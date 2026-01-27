package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"chatbot/config"
	_ "chatbot/docs"
	"chatbot/internal/controller/http/handler"
	middleware "chatbot/internal/controller/http/middlerware"

	// middleware "chatbot/internal/controller/http/middlerware"
	"chatbot/internal/usecase"
	"chatbot/pkg/minio"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// NewRouter -.
// Swagger spec:
// @title       Chatbot API
// @description This is a sample server Chatbot server.
// @version     1.0
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func NewRouter(engine *gin.Engine, config *config.Config, useCase *usecase.UseCase, gemini_client *genai.Client, rdb *redis.Client, minioClient *minio.MinIO) {
	// Options
	engine.Use(gin.Logger())
	// engine.Use(gin.Recovery())

	handlerV1 := handler.NewHandler(config, useCase, gemini_client, rdb, *minioClient)
	// Initialize Casbin enforcer

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5050",
			"https://ai-1009.ccenter.uz",
			"https://back-ai.center.uz",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.Use(TimeoutMiddleware(5 * time.Minute))

	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	// K8s probe
	engine.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	// Prometheus metrics
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	enforcer, err := casbin.NewEnforcer("./internal/controller/http/casbin/model.conf", "./internal/controller/http/casbin/policy.csv")
	if err != nil {
		slog.Error("Error while creating enforcer: ", "err", err)
	}

	if enforcer == nil {
		slog.Error("Enforcer is nil after initialization!")
		} else {
			slog.Info("Enforcer initialized successfully.")
		}
		
		// Routes
		
		// engine.GET("/responce/list", handlerV1.GetAllChats)
		// engine.POST("/chats/accept", handlerV1.AcceptResponse)
	engine.GET("/ws/:chat_room_id", handlerV1.ChatWS)
	
	engine.POST("/users/login", handlerV1.Login)
	engine.POST("/users/verify", handlerV1.Verify)

	restrictions := engine.Group("/restrictions")
	{
		restrictions.GET("/get", handlerV1.GetRestrictionByID)
		restrictions.GET("/list", handlerV1.GetAllRestrictions)
		restrictions.PUT("/update", handlerV1.UpdateRestriction)
	}

	engine.Use(
		middleware.Identity(handlerV1.UseCase.UserRepo),
		middleware.Authorize(enforcer),
	)

	engine.POST("/img-upload", handlerV1.UploadFile)

	users := engine.Group("/users")
	{
		users.GET("/profile", handlerV1.GetByIdUser)
		users.GET("/list", handlerV1.GetAllUsers)
		// users.POST("/register", handlerV1.Register)
		users.PUT("/update", handlerV1.UpdateUser)
		users.DELETE("/delete", handlerV1.DeleteUser)
		users.POST("/logout", handlerV1.Logout)
		users.GET("/me", handlerV1.GetMe)
	}


	chat := engine.Group("/chat")
	{
		chat.POST("/room/create", handlerV1.CreateChatRoom)
		chat.DELETE("/room/delete", handlerV1.DeleteChatRoom)
		chat.GET("/user_id", handlerV1.GetChatRoomsByUserId)
		chat.GET("/message", handlerV1.GetChatRoomChat)
	}

	// dashboard := engine.Group("/dashboard")
	// {
	// 	dashboard.GET("/active-users", handlerV1.DashboardActiveUsers)
	// }
}
