package helper

import (
	"bytes"
	"chatbot/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
)

type SmsRequest struct {
	MobilePhone string `json:"mobile_phone"`
	Message     string `json:"message"`
	From        string `json:"from"`
	CallbackURL string `json:"callback_url,omitempty"`
}

func SendSms(cnf config.Config, phone, code string) {
	url := "https://notify.eskiz.uz/api/message/sms/send"

	message := "tasdiqlash kodi - " + code

	smsRequest := SmsRequest{
		MobilePhone: phone,
		Message:     message,
		From:        "4546",
	}

	jsonData, err := json.Marshal(smsRequest)
	if err != nil {
		slog.Error("Error marshaling JSON: ", "err", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Error creating request: ", "err", err)
	} else {
		fmt.Println("Request to Eskiz.uz")
		fmt.Println(req.Body)
		fmt.Println("Request to Eskiz.uz")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cnf.SMS_TOKEN.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request: ", "err", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response body: ", "err", err)
	} else {
		fmt.Println("Response from Eskiz.uz")
		fmt.Println(string(body))
		fmt.Println("Response from Eskiz.uz")
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to send SMS", "status", resp.Status, "body", string(body))
	} else {
		slog.Info("SMS sent successfully", "status", resp.Status)
	}
}
