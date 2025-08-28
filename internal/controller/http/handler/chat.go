package handler

import (
	"context"
	"chatbot/internal/entity"
	"log/slog"

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
// @Param id query string true "User ID"
// @Success 200 {object} entity.ChatRoomList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/user_id [get]
func (h *Handler) GetChatRoomsByUserId(c *gin.Context) {

	res, err := h.UseCase.ChatRepo.GetChatRoomByUserId(context.Background(), &entity.ById{Id: c.Query("id")})
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
// @Success 200 {object} entity.ChatList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /chat/message [get]
func (h *Handler) GetChatRoomChat(c *gin.Context) {

	res, err := h.UseCase.ChatRepo.GetChatRoomChat(context.Background(), &entity.ById{Id: c.Query("id")})
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
