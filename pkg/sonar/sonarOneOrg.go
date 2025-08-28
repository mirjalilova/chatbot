package sonar

import (
	"bufio"
	"bytes"
	"chatbot/internal/entity"
	"chatbot/internal/usecase"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type ppStreamChunk struct {
	ID        string   `json:"id"`
	Object    string   `json:"object"`
	Model     string   `json:"model"`
	Citations []string `json:"citations,omitempty"`

	SearchResults []struct {
		Title string  `json:"title"`
		URL   string  `json:"url"`
		Date  *string `json:"date,omitempty"`
	} `json:"search_results,omitempty"`

	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		Message *struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message,omitempty"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`

	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func StreamToWSOneOrg(db *usecase.UseCase, conn *websocket.Conn, userQuestion, geminiQuestion, chatRoomId string) error {
	fmt.Println("Processing request for one organization (SSE stream)...")

	payload := map[string]any{
		"model": "sonar",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userQuestion},
		},
		"web_search_options": map[string]any{
			"user_location":       map[string]string{"country": "UZ"},
			"search_context_size": "high",
			"search_domain_filter": []string{
				".uz", "www.yellowpages.uz", "www.goldenpages.uz", "https://orginfo.uz/",
			},
		},
		"stream": true,
	}

	req, _ := http.NewRequest("POST", pplxAPIURL, bytes.NewBuffer(mustJSON(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pplxAPIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "text/event-stream") {
		return handleNonStream(db, conn, resp.Body, userQuestion, geminiQuestion, chatRoomId)
	}

	scanner := bufio.NewScanner(resp.Body)
	const maxBuf = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxBuf)

	var fullText string
	citeSeen := map[string]struct{}{}
	var citations []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") ||
			strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "id:") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk ppStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			fmt.Println("WARN: failed to unmarshal SSE chunk:", err)
			continue
		}
		if chunk.Error != nil {
			return fmt.Errorf("sonar stream error: %s", chunk.Error.Message)
		}

		for _, u := range chunk.Citations {
			if _, ok := citeSeen[u]; !ok && strings.TrimSpace(u) != "" {
				citeSeen[u] = struct{}{}
				citations = append(citations, u)
			}
		}
		for _, sr := range chunk.SearchResults {
			if _, ok := citeSeen[sr.URL]; !ok && strings.TrimSpace(sr.URL) != "" {
				citeSeen[sr.URL] = struct{}{}
				citations = append(citations, sr.URL)
			}
		}

		for _, ch := range chunk.Choices {
			if s := ch.Delta.Content; s != "" {
				fullText += s
				_ = conn.WriteJSON(map[string]any{
					"text": s,
				})
			}
		}

	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("stream read error: %v", err)
	}

	_ = conn.WriteJSON(map[string]any{
		"data": map[string]any{
			"text":      fullText,
			"citations": citations,
		},
	})

	go SaveResponce(db, userQuestion, chatRoomId, fullText, geminiQuestion, citations)

	return nil
}

func handleNonStream(db *usecase.UseCase, conn *websocket.Conn, body io.Reader, userQuestion, geminiQuestion, chatRoomId string) error {
	all, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(all, &raw); err != nil {
		return fmt.Errorf("non-stream parse error: %v; body: %s", err, string(all))
	}

	choices, _ := raw["choices"].([]any)
	var text string
	if len(choices) > 0 {
		if msg, ok := choices[0].(map[string]any)["message"].(map[string]any); ok {
			text, _ = msg["content"].(string)
		}
	}
	var citations []string
	if cits, ok := raw["citations"].([]any); ok {
		for _, c := range cits {
			if s, ok := c.(string); ok {
				citations = append(citations, s)
			}
		}
	}
	if srs, ok := raw["search_results"].([]any); ok {
		for _, v := range srs {
			if m, ok := v.(map[string]any); ok {
				if url, ok := m["url"].(string); ok {
					citations = append(citations, url)
				}
			}
		}
	}

	_ = conn.WriteJSON(map[string]any{
		"data": map[string]any{
			"text":      text,
			"citations": citations,
		},
	})

	go SaveResponce(db, userQuestion, chatRoomId, text, geminiQuestion, citations)

	return nil
}

/*************  ✨ Windsurf Command ⭐  *************/
// SaveResponce saves a chat log with the given request, chat room ID, response,
// Gemini request, and citation URLs.
/*******  7e881605-fe22-4492-bdeb-47cb86f30523  *******/
func SaveResponce(db *usecase.UseCase, request, chat_room_id, responce, gemini_request string, citation_urls []string) {

	db.ChatRepo.Create(context.Background(), &entity.ChatCreate{
		ChatRoomID:    chat_room_id,
		UserRequest:   request,
		GeminiRequest: gemini_request,
		Responce:      responce,
		CitationURLs:  citation_urls,
	})

	fmt.Println("Saving chat log:", request, responce)
}
