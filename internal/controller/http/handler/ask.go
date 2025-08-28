package handler

// import (
// 	"chatbot/internal/entity"
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"net/http"
// 	"time"

// 	"chatbot/pkg/cache"
// 	"chatbot/pkg/gemini"

// 	"chatbot/pkg/sonar"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt"
// 	"github.com/google/uuid"
// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool { return true },
// }

// // Ask godoc
// // @Summary Send a message to OpenAI Assistant
// // @Description Sends a user message to the fine-tuned GPT assistant and returns the assistant's reply
// // @Tags Assistant
// // @Accept json
// // @Produce json
// // @Param request body entity.AskRequest true "Message to Assistant"
// // @Success 200 {object} entity.AskResponse
// // @Failure 400 {object} map[string]string
// // @Failure 500 {object} map[string]string
// // @Security BearerAuth
// // @Router /ask [post]
// func (h *Handler) Ask(c *gin.Context) {

// 	r := gin.Default()
// 	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		fmt.Println("WebSocket upgrade error:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	var ctx = context.Background()
// 	var req entity.AskRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	claims, exists := c.Get("claims")
// 	if !exists {
// 		slog.Error("error", "", "Unauthorized")
// 		c.JSON(401, entity.ErrorResponse{
// 			Message: "Unauthorized",
// 		})
// 		return
// 	}

// 	mapClaims, ok := claims.(jwt.MapClaims)
// 	if !ok {
// 		c.JSON(401, gin.H{"message": "Invalid token claims"})
// 		return
// 	}

// 	user_id, ok := mapClaims["id"].(string)
// 	if !ok {
// 		c.JSON(400, gin.H{"message": "Invalid user ID type in token"})
// 		return
// 	}

// 	oldQueries, err := cache.GetUserQueries(h.Redis, ctx, user_id, int64(5))

// 	geminiResp := gemini.GetResponse(req.Request, oldQueries)
// 	chat_id := uuid.New().String()

// 	if geminiResp.Route == "gemini" {
// 		r.GET(fmt.Sprintf("/ws/%s", chat_id), func(c *gin.Context) {

// 			go func() {
// 				if err := cache.AppendUserQuery(h.Redis, ctx, "12345678", req.Request); err != nil {
// 					slog.Warn("Failed to append user query", "error", err)
// 				}
// 			}()

// 			_ = conn.WriteJSON(map[string]any{
// 				"responce": geminiResp.Explanation,
// 			})
// 			slog.Info("Gemini response", "explanation", geminiResp.Explanation)
// 			return
// 		})
// 	}

// 	if geminiResp.ExpectsMultiple {
// 		resp, err := sonar.StreamToWS(geminiResp.EnrichedQuery)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Sonar error: %v", err)})
// 			return
// 		}
// 		c.JSON(http.StatusOK, resp)
// 		return
// 	} else {
// 		if err := sonar.StreamToWSOneOrg(conn, geminiResp.EnrichedQuery); err != nil {
// 			slog.Error("Sonar error", "error", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Sonar error: %v", err)})
// 			return
// 		}
// 	}

// 	err = h.UseCase.ChatRepo.Check(context.Background(), user_id, req.ChatRoomID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		slog.Error("failed to check chat room: ", "err", err)
// 		return
// 	}

// 	threadID, err := h.UseCase.ChatRepo.GetThreadID(context.Background(), req.ChatRoomID)
// 	if err != nil {
// 		if err.Error() == "no rows in result set" {
// 			threadID = nil
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			slog.Error("failed to get thread ID: ", "err", err)
// 			return
// 		}
// 	}

// 	var threadIDVal string
// 	if threadID != nil {
// 		threadIDVal = *threadID
// 	} else {
// 		threadIDVal = ""
// 	}

// 	id := uuid.New().String()
// 	go saveResponce(h, geminiResp.EnrichedQuery, req.ChatRoomID, chat_id, threadIDVal)

// 	c.JSON(http.StatusOK, gin.H{
// 		"id":      id,
// 		"message": geminiResp.Explanation,
// 	})
// }

// // GetResponce godoc
// // @Summary Get a responce from OpenAI Assistant
// // @Description Get a responce from OpenAI Assistant
// // @Tags Assistant
// // @Accept json
// // @Produce json
// // @Param id query string true "Chat ID"
// // @Success 200 {object} entity.AskResponse
// // @Failure 400 {object} map[string]string
// // @Failure 500 {object} map[string]string
// // @Security BearerAuth
// // @Router /get/gpt/responce [get]
// func (h *Handler) GetResponce(c *gin.Context) {

// 	answer, err := h.UseCase.ChatRepo.GetById(context.Background(), &entity.ById{
// 		Id: c.Query("id"),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get response"})
// 		slog.Error("failed to get response: ", "err", err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"responce": answer.Answer})
// }

// func saveResponce(h *Handler, request, chat_room_id, id, threadID string) {

// 	if threadID == "" {

// 		// 1. Create thread
// 		threadResp := struct {
// 			ID string `json:"id"`
// 		}{}
// 		_, err := h.OpenAIClient.Client.R().
// 			SetResult(&threadResp).
// 			Post(h.OpenAIClient.BaseURL + "/threads")
// 		if err != nil {
// 			slog.Error("failed to create thread: ", "err", err)
// 			return
// 		}
// 		threadID = threadID
// 	}

// 	// 2. Add message
// 	_, err := h.OpenAIClient.Client.R().
// 		SetBody(map[string]interface{}{
// 			"role":    "user",
// 			"content": request,
// 		}).
// 		Post(h.OpenAIClient.BaseURL + "/threads/" + threadID + "/messages")
// 	if err != nil {
// 		slog.Error("failed to send message: ", "err", err)
// 		return
// 	}

// 	// 3. Run assistant
// 	runResp := struct {
// 		ID string `json:"id"`
// 	}{}
// 	_, err = h.OpenAIClient.Client.R().
// 		SetBody(map[string]interface{}{
// 			"assistant_id": h.OpenAIClient.AssistantID,
// 		}).
// 		SetResult(&runResp).
// 		Post(h.OpenAIClient.BaseURL + "/threads/" + threadID + "/runs")
// 	if err != nil {
// 		slog.Error("failed to start assistant run: ", "err", err)
// 		return
// 	}

// 	// 4. Wait for run to complete
// 	for {
// 		status := struct {
// 			Status string `json:"status"`
// 		}{}
// 		_, err := h.OpenAIClient.Client.R().
// 			SetResult(&status).
// 			Get(h.OpenAIClient.BaseURL + "/threads/" + threadID + "/runs/" + runResp.ID)
// 		if err != nil {
// 			slog.Error("failed to check run status: ", "err", err)
// 			return
// 		}
// 		if status.Status == "completed" {
// 			break
// 		}
// 		time.Sleep(1 * time.Second)
// 	}

// 	// 5. Get assistant response
// 	msgResp := struct {
// 		Data []struct {
// 			Role    string `json:"role"`
// 			Content []struct {
// 				Text struct {
// 					Value string `json:"value"`
// 				} `json:"text"`
// 			} `json:"content"`
// 		} `json:"data"`
// 	}{}

// 	_, err = h.OpenAIClient.Client.R().
// 		SetResult(&msgResp).
// 		Get(h.OpenAIClient.BaseURL + "/threads/" + threadID + "/messages")
// 	if err != nil {
// 		slog.Error("failed to get response: ", "err", err)
// 		return
// 	}

// 	answer := ""
// 	for _, m := range msgResp.Data {
// 		if m.Role == "assistant" && len(m.Content) > 0 {
// 			answer = m.Content[0].Text.Value
// 			break
// 		}
// 	}

// 	h.UseCase.ChatRepo.Create(context.Background(), &entity.ChatCreate{
// 		Id:         id,
// 		ChatRoomID: chat_room_id,
// 		Request:    request,
// 		Responce:   answer,
// 		ThreadID:   threadID,
// 	})

// 	fmt.Println("Saving chat log:", id, request, answer)
// }
