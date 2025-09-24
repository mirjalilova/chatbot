package entity

type ChatCreate struct {
	Id            string   `json:"id"`
	ChatRoomID    string   `json:"chat_room_id" binding:"required"`
	UserRequest   string   `json:"user_request"`
	GeminiRequest string   `json:"gemini_request"`
	Responce      string   `json:"responce" binding:"required"`
	Location      []string `json:"location" binding:"required"`
	ImagesURL     []string `json:"images_url" binding:"required"`
	Organizations any      `json:"organizations" binding:"required"`
	CitationURLs  []string `json:"citation_urls" binding:"required"`
}

type ChatRoomCreate struct {
	UserId string `json:"user_id" binding:"required"`
}

type ChatRoom struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

type ChatRoomList struct {
	ChatRooms []ChatRoom `json:"chat_rooms"`
	Count     int        `json:"count"`
}

type Chat struct {
	ID         string  `json:"id"`
	ChatRoomID string  `json:"chat_room_id"`
	Role       string  `json:"role"`
	Content    Content `json:"content"`
	CreatedAt  string  `json:"created_at"`
}

type Content struct {
	Text          string   `json:"text,omitempty"`
	Citations     []string `json:"citations,omitempty"`
	Location      []string `json:"location,omitempty"`
	ImagesURL     []string `json:"images_url,omitempty"`
	Organizations any      `json:"organizations,omitempty"`
}

type ChatList struct {
	Chats []Chat `json:"chats"`
}

type AskResponse struct {
	Answer string `json:"responce"`
}

type MessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type MessageResponse struct {
	Reply string `json:"reply"`
}

type OrgInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Phone       string   `json:"phone,omitempty"`
	Email       string   `json:"email,omitempty"`
	Description string   `json:"description,omitempty"`
	Website     string   `json:"website,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	ImagesURL   []string `json:"images_url,omitempty"`
}

type Response struct {
	Citations []string  `json:"citations"`
	Data      []OrgInfo `json:"data"`
}

type Request struct {
	Message string `json:"message"`
}

type AskRequest struct {
	Message string `json:"message" binding:"required"`
}

type GetChatRoomReq struct {
	UserId string `json:"user_id" binding:"required"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
