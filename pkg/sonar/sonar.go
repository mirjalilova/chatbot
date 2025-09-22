package sonar

import (
	"bytes"
	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/internal/usecase"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	pplxAPIURL = "https://api.perplexity.ai/chat/completions"
)

var systemPrompt = `
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
Return the answer only in the language of the question given
`

func StreamToWS(cfg config.Config, db *usecase.UseCase, conn *websocket.Conn, userQuestion, geminiQuestion, chatRoomId string) error {
	payload := map[string]any{
		"model": "sonar",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": geminiQuestion},
		},
		"web_search_options": map[string]any{
			"user_location":       map[string]string{"country": "UZ"},
			"search_context_size": "medium",
			"search_domain_filter": []string{
				".uz", "www.yellowpages.uz", "www.goldenpages.uz", "https://orginfo.uz/",
			},
		},
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"schema": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name":    map[string]string{"type": "string"},
							"address": map[string]string{"type": "string"},
							"location": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"latitude":  map[string]string{"type": "number"},
									"longitude": map[string]string{"type": "number"},
								},
								"required": []string{"latitude", "longitude"},
							},
							"phone":       map[string]string{"type": "string"},
							"email":       map[string]string{"type": "string"},
							"description": map[string]string{"type": "string"},
							"website":     map[string]string{"type": "string"},
							"sources": map[string]any{
								"type":  "array",
								"items": map[string]string{"type": "string"},
							},
							"images_url": map[string]any{
								"type":  "array",
								"items": map[string]string{"type": "string"},
							},
						},
						"required": []string{"name", "address"},
					},
				},
			},
		},

		"stream": false,
	}

	// _ = conn.WriteJSON(map[string]any{
	// 	"type":    "payload",
	// 	"payload": payload,
	// })

	req, _ := http.NewRequest("POST", pplxAPIURL, bytes.NewBuffer(mustJSON(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.PerplexityAPIKey.Key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}

	choices, ok := raw["choices"].([]any)
	if !ok || len(choices) == 0 {
		return fmt.Errorf("no choices returned from Sonar")
	}

	citationsAny, ok := raw["citations"].([]any)
	if !ok {
		return fmt.Errorf("invalid citations format from Sonar")
	}

	citations := make([]string, 0, len(citationsAny))
	for _, c := range citationsAny {
		s, ok := c.(string)
		if !ok {
			return fmt.Errorf("citation is not a string: %v", c)
		}
		citations = append(citations, s)
	}

	msg := choices[0].(map[string]any)["message"].(map[string]any)
	content, ok := msg["content"].(string)
	if !ok {
		return fmt.Errorf("invalid response format from Sonar")
	}

	var orgs []entity.OrgInfo
	if err := json.Unmarshal([]byte(content), &orgs); err != nil {
		return fmt.Errorf("failed to parse structured content: %v", err)
	}

	fmt.Println("\n\n\nSonar response:", orgs)

	res := &entity.Response{
		Citations: citations,
		Data:      orgs,
	}

	err = conn.WriteJSON(map[string]any{
		"response": res,
	})
	if err != nil {
		return err
	}

	err = conn.WriteJSON(map[string]any{
		"status": "end",
	})
	if err != nil {
		return err
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return err
	}

	go SaveResponce(db, userQuestion, chatRoomId, string(resBytes), geminiQuestion, citations)

	return nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
