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
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		var req entity.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			fmt.Println("json parse error:", err)
			continue
		}
		request := req.Message

		go func() {
			if err := cache.AppendUserQuery(h.Redis, ctx, "12345678", request); err != nil {
				slog.Warn("Failed to append user query", "error", err)
			}
		}()

		oldQueries, err := cache.GetUserQueries(h.Redis, ctx, "12345678", int64(5))

		geminiResp := gemini.GetResponse(*h.Config, request, oldQueries)

		if geminiResp.Route == "gemini" {
			err = conn.WriteJSON(map[string]any{
				"response": geminiResp.Explanation,
			})
			if err != nil {
				fmt.Println("write error:", err)
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
			if err := sonar.StreamToWSOneOrg(*h.Config, h.UseCase, conn, request, geminiResp.EnrichedQuery, chatRoomID); err != nil {
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
	})

	fmt.Println("Saving chat log:", request, responce)
}
