package handler

// import (
// 	"chatbot/config"
// 	"chatbot/internal/entity"
// 	"log/slog"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// // DashboardActiveUsers gets the number of active users in a given time range
// // @Summary Get active users count
// // @Description Get the count of active users between two dates
// // @Tags Dashboard
// // @Accept  json
// // @Produce  json
// // @Param fromDate query string true "From Date"
// // @Param toDate query string true "To Date"
// // @Success 200 {object} []entity.DashboardActiveUsers
// // @Failure 400 {object} entity.ErrorResponse
// // @Failure 500 {object} entity.ErrorResponse
// // @Router /dashboard/active-users [get]
// func (h *Handler) DashboardActiveUsers(ctx *gin.Context) {
// 	fromDate, err := time.Parse("2006-01-02", ctx.Query("fromDate"))
// 	if err != nil {
// 		slog.Error("GetAudioTranscriptStats error", slog.String("error", err.Error()))
// 		ctx.JSON(400, entity.ErrorResponse{
// 			Code:    config.ErrorBadRequest,
// 			Message: "Invalid fromDate format, expected YYYY-MM-DD",
// 		})
// 		return
// 	}
// 	toDate, err := time.Parse("2006-01-02", ctx.Query("toDate"))
// 	if err != nil {
// 		slog.Error("GetAudioTranscriptStats error", slog.String("error", err.Error()))
// 		ctx.JSON(400, entity.ErrorResponse{
// 			Code:    config.ErrorBadRequest,
// 			Message: "Invalid toDate format, expected YYYY-MM-DD",
// 		})
// 		return
// 	}

// 	res, err := h.UseCase.DashboardRepo.GetUserAndRequestCount(ctx, fromDate, toDate)
// 	if err != nil {
// 		slog.Error("GetUserAndRequestCount error", slog.String("error", err.Error()))
// 		ctx.JSON(500, entity.ErrorResponse{
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	ctx.JSON(200, res)
// }
