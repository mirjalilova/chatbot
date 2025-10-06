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
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
)

func OrganizationCreate(cfg config.Config, r redis.Client, sonarResp string, organizations []cache.Organization, chatRoomId string) []cache.Organization {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.ApiKey.Key))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil
	}
	defer client.Close()

	model := client.GenerativeModel("models/gemini-2.5-pro")
	advice := model.StartChat()

	prompt := fmt.Sprintf(`
You are an intelligent assistant that manages and updates organization information.

You are given:
1️⃣ **Sonar Response** – contains data about one organization (either new or updated).
2️⃣ **Organizations** – the current array of already stored organizations.

🎯 Your task:
- Convert the Sonar response into the correct JSON structure (organization format below).
- If an organization with a similar or same **name** already exists, update its fields:
  - Add or update missing information.
  - Keep any old data that is not mentioned in the new response.
- If it’s a new organization, append it to the list.
- The organization whose information is currently being received should always appear as the **first element** of the array (index 0).
- Do not remove or lose any existing organization data.

---

📘 **Output Format**
Return a valid JSON **array** of all organizations, following this structure:

[
  {
    "name": "Tech Innovators Inc.",
    "description": "A global technology company specializing in AI and cloud solutions.",
    "founded_year": 2010,
    "industry": "Information Technology",
    "headquarters": {
      "address": "123 Innovation Drive",
      "city": "San Francisco",
      "country": "USA"
    },
    "contacts": {
      "phone": "+1-415-555-1234",
      "email": "info@techinnovators.com",
      "website": "https://www.techinnovators.com",
      "social_media": {
        "linkedin": "https://www.linkedin.com/company/tech-innovators",
        "twitter": "https://twitter.com/tech_innovators",
        "instagram": "string",
        "telegram": "string"
      }
    },
    "key_people": [
      {
        "name": "Jane Doe",
        "role": "Chief Executive Officer",
        "email": "jane.doe@techinnovators.com"
      }
    ],
    "subsidiaries": [
      {
        "name": "Innovate Labs",
        "industry": "Research & Development",
        "location": "Boston, MA, USA"
      }
    ],
    "registration": {
      "tax_id": "94-1234567",
      "registration_number": "CA-987654321"
    }
  }
]

---

🧠 **Sonar Response:**
%s

📦 **Existing Organizations:**
%s

Return only a clean JSON array of organizations.  
Do not include explanations, text, or markdown fences 
`, sonarResp, organizations)

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

	var parsed []cache.Organization
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		slog.Error("failed to parse Gemini response as JSON", "error", err, "response", clean)
		return nil
	}

	go cache.AppendChatOrganization(&r, context.Background(), "o"+chatRoomId, parsed)

	return parsed
}
