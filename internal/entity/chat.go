package entity

type ChatCreate struct {
	Id            string   `json:"id"`
	ChatRoomID    string   `json:"chat_room_id" binding:"required"`
	UserRequest   string   `json:"user_request"`
	GeminiRequest string   `json:"gemini_request"`
	Responce      string   `json:"responce" binding:"required"`
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
	ID string `json:"id"`
	// UserID     string `json:"user_id"`
	ChatRoomID string `json:"chat_room_id"`
	Request    string `json:"request"`
	Response   string `json:"response"`
	CreatedAt  string `json:"created_at"`
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
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Phone       string   `json:"phone,omitempty"`
	Email       string   `json:"email,omitempty"`
	Description string   `json:"description,omitempty"`
	Website     string   `json:"website,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}

type Response struct {
	Citations []string  `json:"citations"`
	Data      []OrgInfo `json:"data"`
}

type Request struct {
	Message string `json:"message"`
}
