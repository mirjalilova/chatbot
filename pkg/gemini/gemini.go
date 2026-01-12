package gemini

import (
	"chatbot/config"
	"chatbot/pkg/cache"
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

func GetResponse(cfg config.Config, userQuestion string, oldQueries []string, organizations []cache.Organization) *GeminiResponse {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.ApiKey.Key))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil
	}
	defer client.Close()

	historyContext := ""
	if len(oldQueries) > 0 {
		start := 0
		if len(oldQueries) > 5 {
			start = len(oldQueries) - 5
		}
		lastFive := oldQueries[start:]
		historyContext = strings.Join(lastFive, "\n- ")
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	advice := model.StartChat()

	prompt := fmt.Sprintf(`
You are an intelligent query analyzer working as a gateway before sending user questions to the Sonar model.

Your task is to decide whether the user question should be handled by Gemini itself or sent to Sonar for organization-related data retrieval.

---

### 🔹 Responsibilities

1. **Classification**
	 If the user's question is about greetings or what you can do for them, Introduce yourself to him, that is, tell him in detail how you can answer questions about Uzbekistan organizations or share information about them with him:
     {
       "route": "gemini",
       "explanation": "your answer"
     }
   - If the user’s question is **not related to organizations in Uzbekistan**, Politely explain to the user that they are asking about Uzbek organizations and that you cannot answer questions in this area:
     {
       "route": "gemini",
       "explanation": "your answer"
     }
   - If the question **is related** to organizations in Uzbekistan (such as company name, address, contact, ranking, size, or type), continue to step 2.

---

2. **Enrichment and Context Understanding**
   - Rephrase and enrich the question to make it more complete for Sonar.
   - You are provided with:
     - Conversation history (the last 5 user queries).
     - A list of known organizations, where the **0-index organization** is the most recently discussed or most relevant one.
   - If the current question is short or refers to words like “it”, “they”, or uses implicit references such as “address?”, “phone number?”, “what about it?”, assume it refers to the **0-index organization** in the list.
   - When enriching the question, include:
     - The organization name from the 0-index if applicable.
     - Relevant context such as organization type, location (Uzbekistan), and what information the user might be seeking (e.g., ranking, contacts, description).

---

3. **Multiplicity prediction**
   - Predict if the question expects information about multiple organizations or just one.
   - Return this as "expects_multiple":
     - true → if the question is about categories, lists, or comparisons (e.g., "the biggest universities in Uzbekistan").
     - false → if the question is about a single specific organization.

---

4. **Return Format**
   - Always return valid JSON only (no extra explanations, markdown, or text).
   - Example for non-organization questions:
     {
       "route": "gemini",
       "explanation": "This question is about weather, not organizations."
     }

   - Example for organization-related questions:
     {
       "route": "sonar",
       "enriched_query": "What is the address and contact number of Tashkent University of Information Technologies in Uzbekistan?",
       "expects_multiple": false
     }

**Rules:**
- Always return valid JSON (no extra text outside of JSON).
- "route" is always either "gemini" or "sonar".
- If "route" is "gemini", include "explanation".
- If "route" is "sonar", include "enriched_query" and "expects_multiple" (true/false).

---

🧠 **Conversation history (last 5 questions):**
- %s

🏢 **Known organizations:**
%s

💬 **Current user question:**
%s

Reply **only in valid JSON** and in **Uzbek language**.
`, historyContext, organizations, userQuestion)

	res, err := advice.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		slog.Error("failed to send message to Gemini", "error", err)
		return nil
	}

	var builder strings.Builder
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			builder.WriteString(fmt.Sprintf("%v", part))
		}
	}

	raw := strings.TrimSpace(builder.String())
	clean := strings.Trim(raw, "` \n")
	clean = strings.TrimPrefix(clean, "json")
	clean = strings.TrimPrefix(clean, "JSON")

	var parsed GeminiResponse
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		slog.Error("failed to parse Gemini response as JSON", "error", err)
		return nil
	}

	// if parsed.Route == "gemini" {
	// 	fmt.Println("Return directly to user:", parsed.Explanation)
	// } else if parsed.Route == "sonar" {
	// 	fmt.Println("Forward to Sonar with enriched query:", parsed.EnrichedQuery)
	// }

	return &parsed
}
