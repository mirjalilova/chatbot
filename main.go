package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"chatbot/pkg/cache"
	"chatbot/pkg/gemini"
	"chatbot/pkg/sonar"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }}

func main() {

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "4545",
		DB:       0,
	})

	r := gin.Default()

	r.Static("/assets", "./public")
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WS read error:", err)
			return
		}
		userQuestion := string(msg)

		go func() {
			if err := cache.AppendUserQuery(rdb, ctx, "12345678", userQuestion); err != nil {
				slog.Warn("Failed to append user query", "error", err)
			}
		}()

		oldQueries, err := cache.GetUserQueries(rdb, ctx, "12345678", int64(5))

		geminiResp := gemini.GetResponse(userQuestion, oldQueries)

		if geminiResp.Route == "gemini" {
			_ = conn.WriteJSON(map[string]any{
				"responce": geminiResp.Explanation,
			})
			return
		}

		systemPrompt := `
Respond to user queries by retrieving and presenting information on organizations in Uzbekistan only from reliable, verifiable sources (e.g., official government registries, reputable news outlets, recognized business directories, or accredited databases).

Response Guidelines:

Source Reliability: Only provide information if it can be verified by at least one reliable source.

Transparency: Always cite the source(s) in your response.

No Guesswork: If no reliable source is found, clearly state: "No reliable information available." Do not speculate, fabricate, or infer details.

Geographic Scope: Only return results about organizations physically located in Uzbekistan.

Relevance: Ensure the information directly answers the user's request without unrelated details.

Neutrality: Present information factually and without bias. Avoid opinions or promotional language.

Fail-safe Rule:
If you cannot confirm the accuracy of the information or cannot locate a trustworthy source, you must respond with:
"No reliable information available."
`
		if geminiResp.ExpectsMultiple {
			if err := sonar.StreamToWS(conn, geminiResp.EnrichedQuery, systemPrompt); err != nil {
				_ = conn.WriteJSON(map[string]any{
					"type":  "error",
					"error": fmt.Sprintf("Sonar error: %v", err),
				})
				return
			}
		} else {
			if err := sonar.StreamToWSOneOrg(conn, geminiResp.EnrichedQuery, systemPrompt); err != nil {
				_ = conn.WriteJSON(map[string]any{
					"type":  "error",
					"error": fmt.Sprintf("Sonar error: %v", err),
				})
				return
			}
		}
	})

	_ = r.Run(":8080")
}
