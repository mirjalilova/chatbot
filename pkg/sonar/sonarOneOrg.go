package sonar

import (
	"bufio"
	"bytes"
	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/internal/usecase"
	"chatbot/pkg/coords"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Location *struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location,omitempty"`
}

// var systemPrompt2 = `
// Respond to user queries by retrieving and presenting information on organizations in Uzbekistan only from reliable, verifiable sources (e.g., official registries, business directories, or government databases).

// Response Guidelines:

// 1. **Organization Info**: Provide factual details (name, address, phone number, services, etc.) ONLY if verified by reliable sources.
// 2. **Location**: Determine the location of the organization and return its latitude and longitude. Please note that the location of the organization must be returned and it is important that it is correct.
//    "location": {
//       "lat": <latitude>,
//       "lng": <longitude>
//    }
//    - If exact coordinates are not available, try to approximate using Google Maps or official references.
//    - If no coordinates are available, respond with:
//      "location": null
// 3. **Sources**: Always include citation links for the information you provide.
// 4. **No Guesswork**: Do not invent or speculate details. If not available from reliable sources, clearly state: "No reliable information available."
// 5. **Geographic Scope**: Limit results strictly to organizations physically located in Uzbekistan.
// 6. **Language**: Respond in the same language as the user's question.
// 7. **Output Format**:
//    - Stream partial text as plain chunks for user readability.
//    - At the end, return a final JSON block containing:
//      {
//        "text": "<full answer text>",
//        "citations": [<list of source URLs>],
//        "location": {
//           "lat": <latitude>,
//           "lng": <longitude>
//        }
//      }
// `

func StreamToWSOneOrg(cfg config.Config, db *usecase.UseCase, conn *websocket.Conn, userQuestion, geminiQuestion, chatRoomId string) error {
	fmt.Println("Processing request for one organization (SSE stream)...")

	payload := map[string]any{
		"model": "sonar",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": geminiQuestion},
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
	req.Header.Set("Authorization", "Bearer "+cfg.PerplexityAPIKey.Key)
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

	var locations []map[string]float64

	for _, v := range citations {
		if strings.Contains(v, "maps.google.com") || strings.Contains(v, "google.com/maps") {
			u, err := url.Parse(v)
			if err != nil {
				continue
			}

			if strings.Contains(u.Path, "@") {
				parts := strings.Split(u.Path, "@")
				if len(parts) > 1 {
					coords := strings.Split(parts[1], ",")
					if len(coords) >= 2 {
						lat := coords[0]
						lng := coords[1]
						locations = append(locations, map[string]float64{
							"lat": parseFloat(lat),
							"lng": parseFloat(lng),
						})
					}
				}
			}

			q := u.Query().Get("q")
			if q != "" {
				coords := strings.Split(q, ",")
				if len(coords) >= 2 {
					lat := coords[0]
					lng := coords[1]
					locations = append(locations, map[string]float64{
						"lat": parseFloat(lat),
						"lng": parseFloat(lng),
					})
				}
			}
		} else if strings.Contains(v, "https://yandex.uz/maps/org") || strings.Contains(v, "https://yandex.com/maps/org") {

			fmt.Println(v)
			lat, lng, err := coords.ExtractCoordinates(v)
			fmt.Println("Yandex coords:", lat, lng, "Error:", err)
			if err == nil {
				fmt.Println("Got coords from Yandex:", lat, lng)
				locations = append(locations, map[string]float64{
					"lat": lat,
					"lng": lng,
				})
			}
		}
	}

	var finalLocations []map[string]float64
	if len(locations) > 0 {
		finalLocations = locations
	} else {
		finalLocations = nil
	}

	images := extractImageURLs(citations)

	_ = conn.WriteJSON(map[string]any{
		"data": map[string]any{
			"text":          fullText,
			"citations":     citations,
			"location":      finalLocations,
			"images_url":    images,
			"organizations": nil,
		},
	})

	err = conn.WriteJSON(map[string]any{
		"status": "end",
	})
	if err != nil {
		return err
	}

	go SaveResponce(db, userQuestion, chatRoomId, fullText, geminiQuestion, citations, finalLocations, images, nil)

	return nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f
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
			"text":          text,
			"citations":     citations,
			"location":      nil,
			"images_url":    nil,
			"organizations": nil,
		},
	})

	err = conn.WriteJSON(map[string]any{
		"status": "end",
	})
	if err != nil {
		return err
	}

	go SaveResponce(db, userQuestion, chatRoomId, text, geminiQuestion, citations, nil, nil, nil)

	return nil
}

func SaveResponce(db *usecase.UseCase, request, chat_room_id, responce, gemini_request string, citation_urls []string, locations []map[string]float64, images_url []string, orgs any) {

	db.ChatRepo.Create(context.Background(), &entity.ChatCreate{
		ChatRoomID:    chat_room_id,
		UserRequest:   request,
		GeminiRequest: gemini_request,
		Responce:      responce,
		CitationURLs:  citation_urls,
		Location:      locations,
		ImagesURL:     images_url,
		Organizations: orgs,
	})

	fmt.Println("Saving chat log:", request, responce)
}

func extractImageURLs(urls []string) []string {
	var imgs []string
	for _, u := range urls {
		lower := strings.ToLower(u)
		if strings.HasSuffix(lower, ".jpg") ||
			strings.HasSuffix(lower, ".jpeg") ||
			strings.HasSuffix(lower, ".png") ||
			strings.HasSuffix(lower, ".gif") ||
			strings.HasSuffix(lower, ".webp") {
			imgs = append(imgs, u)
		}
	}
	return imgs
}
