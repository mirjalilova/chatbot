// main.go
package main

// import (
// 	"io"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"

// 	// Swagger
// 	_ "chatbot/test/docs"

// 	swaggerFiles "github.com/swaggo/files"
// 	ginSwagger "github.com/swaggo/gin-swagger"
// )

// type AskRequest struct {
// 	Question string `json:"question"`
// }

// // AskResponse - javob modeli
// type AskResponse struct {
// 	Answer string `json:"answer"`
// }

// type DataItem struct {
// 	Name        string   `json:"name"`
// 	Address     string   `json:"address"`
// 	Phone       string   `json:"phone"`
// 	Email       string   `json:"email"`
// 	Description string   `json:"description"`
// 	Website     string   `json:"website"`
// 	Sources     []string `json:"sources"`
// }

// type Response struct {
// 	Citations []string   `json:"citations"`
// 	Data      []DataItem `json:"data"`
// 	Status    string     `json:"status"`
// }

// type Text struct {
// 	Text string `json:"text"`
// }
// type Responce3 struct {
// 	Data      string   `json:"text"`
// 	Citations []string `json:"citations"`
// }

// // Ask godoc
// // @Summary      Ask a question
// // @Description  Ask a question to the chatbot
// // @Tags         ask
// // @Accept       json
// // @Produce      json
// // @Param        request body AskRequest true "Question"
// // @Success      200 {object} Response
// // @Router       /ask [post]
// func askHandler(c *gin.Context) {
// 	var req AskRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
// 		return
// 	}

// 	if req.Question == "1" {
// 		c.JSON(http.StatusOK, gin.H{"responce": "Assalomu alekum! Qanday yordam bera olaman."})
// 		return
// 	}

// 	responce2 := Response{
// 		Citations: []string{
// 			"https://kun.uz/news/2025/01/15/eng-kop-soliq-tolagan-davlat-va-xususiy-kompaniyalar-ontaligi-malum-qilindi",
// 			"https://gov.uz/oz/soliq/sections/view/16393",
// 			"https://www.uzdaily.uz/uz/ozbekistonning-eng-yirik-soliq-tolovchilari-85-trln-som-soliq-toladi/",
// 			"https://www.goldenpages.uz/uz/rubrics/?Id=1391",
// 			"https://www.gazeta.uz/oz/2025/01/15/top-20/",
// 			"https://b-advice.uz/hotlines",
// 			"https://gov.uz/soliq/news/view/14318",
// 			"https://gov.uz/soliq",
// 		},
// 		Data: []DataItem{
// 			{
// 				Name:        "Soliq qo'mitasi",
// 				Address:     "100011, Toshkent, Shayxontohur tumani, A.Qodiriy ko'chasi, 13A",
// 				Phone:       "+998(71) 244-98-98",
// 				Email:       "info@soliq.uz",
// 				Description: "O'zbekiston Respublikasi Iqtisodiyot va Moliya vazirligi huzuridagi Soliq qo'mitasi.",
// 				Website:     "https://soliq.uz",
// 				Sources:     []string{"[1]", "[3]", "[4]", "[5]"},
// 			},
// 			{
// 				Name:        "Davlat soliq boshqarmasi Toshkent shahar soliq xizmati",
// 				Address:     "100011, Toshkent, Shayxontohur tumani, Abay ko'chasi, 4",
// 				Phone:       "+998(71) 244-54-86",
// 				Email:       "None",
// 				Description: "Davlat soliq boshqarmasi Toshkent shahar soliq xizmati.",
// 				Website:     "None",
// 				Sources:     []string{"[1]", "[4]", "[7]", "[8]"},
// 			},
// 			{
// 				Name:        "Andijon viloyati davlat soliq boshqarmasi",
// 				Address:     "170100, Andijon viloyati, Andijon, Oltinko'l ko'chasi, 1",
// 				Phone:       "+998(74) 223-95-23",
// 				Email:       "None",
// 				Description: "Andijon viloyati davlat soliq boshqarmasi.",
// 				Website:     "None",
// 				Sources:     []string{"[2]", "[6]"},
// 			},
// 		},
// 		Status: "ok",
// 	}

// 	if req.Question == "2" {
// 		c.JSON(http.StatusOK, responce2)
// 		return
// 	}

// 	if req.Question == "3" {
// 		responce3 := Responce3{
// 			Data: "Andijon viloyati davlat soliq boshqarmasining manzili **170100, Andijon viloyati, Andijon, Oltinko'l ko'chasi, 1**dir[1][8]. Telefon raqamlari **74 223 9509** va **74 223 9502**[8]. Ishonch telefoni raqami **74 223 95 23** sifatida ko'rsatilgan[2][5].",
// 			Citations: []string{
// 				"https://yandex.uz/maps/org/211366371489/",
// 				"https://gov.uz/oz/soliq/sections/view/16393",
// 				"https://yandex.uz/maps/org/12434187164/",
// 				"https://data.egov.uz/data/610a72cf1a64fdd0373a8e1a",
// 				"https://gov.uz/oz/soliq/pages/contacts",
// 				"https://www.imv.uz/static/andijon-viloyati",
// 				"https://www.goldenpages.uz/uz/company/?Id=45789",
// 				"https://www.goldenpages.uz/uz/company/?Id=42588",
// 				"https://www.davaktiv.uz/oz/corporate/company/THKF",
// 				"https://gov.uz/soliq",
// 			},
// 		}
// 		// Streaming javob
// 		messages := []string{
// 			"Andijon viloyati davlat ",
// 			"soliq boshqarmasining manzili **170100, ",
// 			"Andijon viloyati, Andijon, Oltinko'l ",
// 			"ko'chasi, 1**dir[1][8]. Telefon raqamlari **74 223 9509**.",
// 			" va **74 223 9502**[8]. Ishonch telefoni raqami ",
// 			"**74 223 95 23** sifatida ko'rsatilgan[2][5].",
// 		}

// 		c.Stream(func(w io.Writer) bool {
// 			for _, msg := range messages {
// 				c.SSEvent("message", Text{Text: msg})
// 				time.Sleep(1 * time.Second)
// 			}
// 			c.SSEvent("message", responce3)
// 			return false
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"answer": "Savolni tushunmadim",
// 	})
// }

// func main() {
// 	r := gin.Default()

// 	// Swagger UI
// 	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	// Routes
// 	r.POST("/ask", askHandler)

// 	r.Run(":8080")
// }
