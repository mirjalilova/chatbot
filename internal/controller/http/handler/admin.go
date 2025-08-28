package handler

import (
	"context"
	"log/slog"
	"net/http"

	"chatbot/internal/entity"

	"github.com/gin-gonic/gin"
)

type AcceptRequest struct {
	ID     string `json:"id" binding:"required"`
	Accept bool   `json:"accept"`
}

// GetRestrictionByID godoc
// @Summary Get restriction by ID
// @Description Get restriction by ID
// @Tags Restrictions
// @Accept  json
// @Produce  json
// @Param id query string true "Restriction ID"
// @Success 200 {object} entity.Restriction
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /restrictions/get [get]
func (h *Handler) GetRestrictionByID(c *gin.Context) {
	res, err := h.UseCase.RestrictionRepo.GetById(context.Background(), &entity.ById{Id: c.Query("id")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Get restriction error", "err", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetAllRestrictions godoc
// @Summary Get all restrictions
// @Description Get list of restrictions
// @Tags Restrictions
// @Accept  json
// @Produce  json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} entity.Restriction
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /restrictions/list [get]
func (h *Handler) GetAllRestrictions(c *gin.Context) {
	limit, offset, err := parsePaginationParams(c, c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Pagination parse error", "err", err)
		return
	}

	res, err := h.UseCase.RestrictionRepo.GetAll(context.Background(), &entity.Filter{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Get all restrictions error", "err", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// UpdateRestriction godoc
// @Summary Update a restriction
// @Description Update an existing restriction
// @Tags Restrictions
// @Accept  json
// @Produce  json
// @Param id query string true "Restriction ID"
// @Param restriction body entity.UpdateRestrictionBody true "Update data"
// @Success 200 {string} string "Restriction updated successfully"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /restrictions/update [put]
func (h *Handler) UpdateRestriction(c *gin.Context) {
	var req entity.UpdateRestrictionBody

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Binding error", "err", err)
		return
	}

	if err := h.UseCase.RestrictionRepo.Update(context.Background(), &entity.UpdateRestriction{
		ID:             c.Query("id"),
		RequestLimit:   req.RequestLimit,
		CharacterLimit: req.CharacterLimit,
		ChatLimit:      req.ChatLimit,
		// TimeLimit:      req.TimeLimit,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Update restriction error", "err", err)
		return
	}

	c.JSON(http.StatusOK, "Restriction updated successfully")
}

// // GetAllChats godoc
// // @Summary Get all chats from JSON file
// // @Description Get list of chat logs saved in JSON
// // @Tags Admin
// // @Accept  json
// // @Produce  json
// // @Success 200 {array} ChatLog
// // @Failure 500 {string} string "Internal server error"
// // @Router /responce/list [get]
// func (h *Handler) GetAllChats(c *gin.Context) {
// 	filename := "./internal/media/chats.json"

// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read chat log file"})
// 		slog.Error("Error reading JSON file", "err", err)
// 		return
// 	}

// 	var logs []ChatLog
// 	if len(data) > 0 {
// 		if err := json.Unmarshal(data, &logs); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON format"})
// 			slog.Error("JSON unmarshal error", "err", err)
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, logs)
// }

// // AcceptResponse godoc
// // @Summary Accept or reject a chat log
// // @Description Accept = do nothing, Reject = delete from JSON
// // @Tags Admin
// // @Accept  json
// // @Produce  json
// // @Param input body AcceptRequest true "Accept or reject"
// // @Success 200 {string} string "Operation successful"
// // @Failure 400 {string} string "Bad request"
// // @Failure 500 {string} string "Internal server error"
// // @Router /chats/accept [post]
// func (h *Handler) AcceptResponse(c *gin.Context) {
// 	var req AcceptRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	filename := "./internal/media/chats.json"

// 	if req.Accept {
// 		c.JSON(http.StatusOK, gin.H{"message": "Accepted"})
// 		return
// 	}

// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read chat log file"})
// 		slog.Error("Error reading JSON file", "err", err)
// 		return
// 	}

// 	var logs []ChatLog
// 	if len(data) > 0 {
// 		if err := json.Unmarshal(data, &logs); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON format"})
// 			slog.Error("JSON unmarshal error", "err", err)
// 			return
// 		}
// 	}

// 	newLogs := make([]ChatLog, 0, len(logs))
// 	for _, log := range logs {
// 		if log.ID != req.ID {
// 			newLogs = append(newLogs, log)
// 		}
// 	}

// 	updatedData, err := json.MarshalIndent(newLogs, "", "  ")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode JSON"})
// 		return
// 	}

// 	if err := os.WriteFile(filename, updatedData, 0644); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save JSON file"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
// }
