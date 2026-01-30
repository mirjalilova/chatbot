package handler

import (
	"chatbot/internal/entity"
	"chatbot/pkg/cache"
	"chatbot/pkg/gemini"
	"chatbot/pkg/sonar"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) ChatWS(c *gin.Context) {
	chatRoomID := c.Param("chat_room_id")
	ctx := context.Background()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade error", "error", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Read message error", "error", err)
			break
		}
		var req entity.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			slog.Error("JSON parse error", "error", err)
			continue
		}
		request := req.Message

		_, err = h.UseCase.ChatRepo.Check(ctx, "", chatRoomID)
		if err != nil {
			if err.Error() == "kunlik limiti tugadi" || err.Error() == "sizning 3 ta bepul so‘rovingiz tugadi, davom etish uchun ro‘yxatdan o‘ting" {
				_ = conn.WriteJSON(map[string]any{
					"type":  "warning",
					"error": err.Error(),
				})
				continue
			}
			slog.Error("Check error", "error", err)
			break
		}

		oldQueries, err := cache.GetUserQueries(h.Redis, ctx, chatRoomID, int64(5))
		if err != nil {
			err = conn.WriteJSON(map[string]any{
						"type":  "error",
						"message": err.Error(),
					})
					if err != nil {
						slog.Warn("Failed to get old queries", "error", err)
						break
					}
				return
		}

		var organizations []cache.Organization
		organizations, err = cache.GetChatOrganizations(
			h.Redis,
			ctx,
			"o"+chatRoomID,
			5,
		)
		if err != nil {
			slog.Warn("Failed to get organizations", "error", err)
			organizations = nil
		}


		fmt.Println("Old queries:", oldQueries)

		geminiResp := gemini.GetResponse(*h.Config, request, oldQueries, organizations)
		if geminiResp == nil {
			_ = conn.WriteJSON(map[string]any{
				"type":  "error",
				"error": "Failed to get response from Gemini",
			})
			continue
		}

		go func() {
			if err := cache.AppendUserQuery(h.Redis, ctx, chatRoomID, geminiResp.EnrichedQuery); err != nil {
				slog.Warn("Failed to append user query", "error", err)
			}
		}()

		if geminiResp.Route == "gemini" {
			err = conn.WriteJSON(map[string]any{
				"content": map[string]any{
					"text":          geminiResp.Explanation,
					"citations":     nil,
					"location":      nil,
					"images_url":    nil,
					"organizations": nil,
				},
			})
			if err != nil {
				slog.Error("Write message error", "error", err)
				break
			}

			err = conn.WriteJSON(map[string]any{
				"status": "end",
			})
			if err != nil {
				slog.Warn("Failed to send end status", "error", err)
				break
			}
			go h.SaveResponce(request, chatRoomID, geminiResp.Explanation, "")
			continue
		}

		if geminiResp.ExpectsMultiple {
			if err := sonar.StreamToWS(*h.Config, h.UseCase, conn, request, geminiResp.EnrichedQuery, chatRoomID); err != nil {
				_ = conn.WriteJSON(map[string]any{
					"type":  "error",
					"error": fmt.Sprintf("Sonar error: %v", err),
				})
				return
			}
		} else {
			if err := sonar.StreamToWSOneOrg(*h.Config, h.UseCase, *h.Redis, conn, request, geminiResp.EnrichedQuery, chatRoomID); err != nil {
				_ = conn.WriteJSON(map[string]any{
					"type":  "error",
					"error": fmt.Sprintf("Sonar error: %v", err),
				})
				return
			}
		}
	}
}

func (h *Handler) SaveResponce(request, chat_room_id, responce, gemini_request string) {

	h.UseCase.ChatRepo.Create(context.Background(), &entity.ChatCreate{
		ChatRoomID:    chat_room_id,
		UserRequest:   request,
		GeminiRequest: gemini_request,
		Responce:      responce,
		CitationURLs:  []string{},
		Location:      []string{},
		ImagesURL:     []string{},
		Organizations: []entity.OrgInfo{},
	})
}
