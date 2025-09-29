package gemini

import (
	"chatbot/config"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiResponse struct {
	Route           string `json:"route"`
	Explanation     string `json:"explanation,omitempty"`
	EnrichedQuery   string `json:"enriched_query,omitempty"`
	ExpectsMultiple bool   `json:"expects_multiple,omitempty"`
}

func GetResponse(cfg config.Config, userQuestion string, oldQueries []string) *GeminiResponse {
	geminiClient, err := genai.NewClient(context.Background(), option.WithAPIKey(cfg.ApiKey.Key))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil
	}

	historyContext := ""
	if len(oldQueries) > 0 {
		start := 0
		if len(oldQueries) > 5 {
			start = len(oldQueries) - 5
		}
		lastFive := oldQueries[start:]
		historyContext = strings.Join(lastFive, "\n- ")
	}

	model := geminiClient.GenerativeModel("models/gemini-2.5-pro")
	advice := model.StartChat()

	prompt := fmt.Sprintf(`
You are an intelligent query analyzer working as a gateway before sending user questions to the Sonar model.

Your responsibilities:

1. **Classification**
   - If the user’s question is NOT related to organizations in Uzbekistan, respond with:
     {
       "route": "gemini",
       "explanation": "Polite and detailed explanation why the question is not about organizations in Uzbekistan."
     }
   - If the question IS about organizations in Uzbekistan (e.g., company size, ranking, address, contact info, industry), continue to step 2.

2. **Enrichment**
   - Rephrase, expand, and enrich the user’s question so it is clearer, more complete, and detailed for Sonar search.
   - Consider not only the current user question, but also the conversation history below. 
   - If the current question is vague or refers to "it", "they", or "also", infer the missing details from the history.
   - Include useful context like: type of organization, location (Uzbekistan-specific), what metric matters (size, revenue, employees, reputation).

3. **Multiplicity prediction**
   - Predict if the query will likely return information about multiple organizations (array) or a single organization (one object).
   - Return this as a boolean field "expects_multiple":
     - true → if the query is about categories, rankings, comparisons, or lists of organizations.
     - false → if the query is about one specific organization.
   - Then respond with:
     {
       "route": "sonar",
       "enriched_query": "<your rewritten and detailed query for Sonar>",
       "expects_multiple": <true|false>
     }

**Rules:**
- Always return valid JSON (no extra text outside of JSON).
- "route" is always either "gemini" or "sonar".
- If "route" is "gemini", include "explanation".
- If "route" is "sonar", include "enriched_query" and "expects_multiple" (true/false).

Conversation history (last 5 user questions):
- %s

Current user question:
%s

Reply only in Uzbek

`, historyContext, userQuestion)

	res, err := advice.SendMessage(context.Background(), genai.Text(prompt))
	if err != nil {
		slog.Error("failed to send message", "error", err)
		return nil
	}

	var builder strings.Builder
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			builder.WriteString(fmt.Sprintf("%v", part))
		}
	}

	raw := strings.TrimSpace(builder.String())

	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")

	var parsed GeminiResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(clean)), &parsed); err != nil {
		slog.Error("failed to parse Gemini response as JSON", "error", err)
		return nil
	}

	if parsed.Route == "gemini" {
		fmt.Println("Return directly to user:", parsed.Explanation)
	} else if parsed.Route == "sonar" {
		fmt.Println("Forward to Sonar with enriched query:", parsed.EnrichedQuery)
	}

	return &parsed
}
