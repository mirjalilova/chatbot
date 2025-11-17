package handler

import (
	middleware "chatbot/internal/controller/http/middlerware"
	"chatbot/internal/entity"
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateChatRoom godoc
// @Summary Create a new chat room
// @Description Create a new chat room with the user id
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param user body entity.ChatRoomCreate true "Chat Details"
// @Success 200 {object} string
// @Failure 400 {object}  string                                                                                                                                       
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/room/create [post]
func (h *Handler) CreateChatRoom(c *gin.Context) {
	reqBody := entity.ChatRoomCreate{}
	err := c.BindJSON(&reqBody)
	if err != nil {
		c.JSON(400, gin.H{"Error binding request body": err.Error()})
		slog.Error("Error binding request body: ", "err", err)
		return
	}

	id, err := h.UseCase.ChatRepo.CreateChatRoom(context.Background(), &reqBody)
	if err != nil {
		c.JSON(500, gin.H{"Error creating chat room:": err.Error()})
		slog.Error("Error creating chat room: ", "err", err)
		return
	}

	slog.Info("New chat room created successfully")
	c.JSON(200, gin.H{"Massage": "chat room created successfully",
		"ID": id})
}

// GetChatRoomsByUserId godoc
// @Summary Get chat room by user ID
// @Description Get chat room by user ID
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} entity.ChatRoomList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/user_id [get]
func (h *Handler) GetChatRoomsByUserId(c *gin.Context) {

	claims, err := middleware.ExtractToken(c.Writer, c.Request, h.UseCase.UserRepo)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token missing or invalid"})
		return
	}

	userID, ok := claims["id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	offset := c.Query("offset")
	limit := c.Query("limit")

	limitValue, offsetValue, err := parsePaginationParams(c, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"Error parsing pagination parameters:": err.Error()})
		slog.Error("Error parsing pagination parameters: ", "err", err)
		return
	}

	req := &entity.GetChatRoomReq{
		UserId: userID,
		Limit:  limitValue,
		Offset: offsetValue,
	}
	res, err := h.UseCase.ChatRepo.GetChatRoomByUserId(context.Background(), req)
	if err != nil {
		c.JSON(500, gin.H{"Error getting Chat rooms by user ID: ": err.Error()})
		slog.Error("Error getting Chat rooms by user ID: ", "err", err)
		return
	}

	slog.Info("Chat rooms retrieved successfully")
	c.JSON(200, res)
}

// GetChatRoomChat godoc
// @Summary Get chat room by ID
// @Description Get chat room by ID
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param id query string true "Chat Room ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} entity.ChatList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/message [get]
func (h *Handler) GetChatRoomChat(c *gin.Context) {

	offset := c.Query("offset")
	limit := c.Query("limit")

	limitValue, offsetValue, err := parsePaginationParams(c, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"Error parsing pagination parameters:": err.Error()})
		slog.Error("Error parsing pagination parameters: ", "err", err)
		return
	}

	res, err := h.UseCase.ChatRepo.GetChatRoomChat(context.Background(), &entity.ById{Id: c.Query("id")}, limitValue, offsetValue)
	if err != nil {
		c.JSON(500, gin.H{"Error getting Chat rooms by ID: ": err.Error()})
		slog.Error("Error getting Chat rooms by ID: ", "err", err)
		return
	}

	slog.Info("Chat rooms chat retrieved successfully")
	c.JSON(200, res)
}

// DeleteChatRoom godoc
// @Summary Delete a chat room
// @Description Delete a chat room by ID
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param id query string true "Chat Room ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/room/delete [delete]
func (h *Handler) DeleteChatRoom(c *gin.Context) {
	err := h.UseCase.ChatRepo.DeleteChatRoom(context.Background(), &entity.ById{Id: c.Query("id")})
	if err != nil {
		c.JSON(500, gin.H{"Error deleting chat room: ": err.Error()})
		slog.Error("Error deleting chat room: ", "err", err)
		return
	}

	slog.Info("Chat room deleted successfully")
	c.JSON(200, gin.H{"Message": "Chat room deleted successfully"})
}
