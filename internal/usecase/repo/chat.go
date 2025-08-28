package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"chatbot/config"
	"chatbot/internal/entity"
	"chatbot/pkg/postgres"
)

type ChatRepo struct {
	pg     *postgres.Postgres
	config *config.Config
}

func NewChatRepo(pg *postgres.Postgres, config *config.Config) *ChatRepo {
	return &ChatRepo{
		pg:     pg,
		config: config,
	}
}

func (r *ChatRepo) CreateChatRoom(ctx context.Context, req *entity.ChatRoomCreate) (string, error) {
	query := `INSERT INTO chat_rooms (user_id) VALUES ($1) RETURNING id`

	var id string
	err := r.pg.Pool.QueryRow(ctx, query, req.UserId).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *ChatRepo) Create(ctx context.Context, req *entity.ChatCreate) error {

	query := `INSERT INTO chat (chat_room_id, user_request, gemini_request, responce,citation_urls) VALUES ($1, $2, $3, $4,$5)`

	_, err := r.pg.Pool.Exec(ctx, query, req.ChatRoomID, req.UserRequest, req.GeminiRequest, req.Responce, req.CitationURLs)
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepo) GetChatRoomByUserId(ctx context.Context, id *entity.ById) (*entity.ChatRoomList, error) {
	query := `
		SELECT COUNT(id) OVER () AS total_count, id, title, created_at
		FROM chat_rooms
		WHERE user_id = $1 AND deleted_at = 0 ORDER BY created_at DESC`

	rows, err := r.pg.Pool.Query(ctx, query, id.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("chat rooms list is empty")
		}
		return nil, err
	}
	defer rows.Close()

	var result entity.ChatRoomList
	for rows.Next() {
		var r entity.ChatRoom
		var createdAt time.Time
		var count int
		err := rows.Scan(&count, &r.ID, &r.Title, &createdAt)
		if err != nil {
			return nil, err
		}

		r.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		result.ChatRooms = append(result.ChatRooms, r)
		result.Count = count

	}

	return &result, nil
}

func (r *ChatRepo) GetChatRoomChat(ctx context.Context, id *entity.ById) (*entity.ChatList, error) {
	query := `
		SELECT id, chat_room_id, user_request, responce, created_at
		FROM chat
		WHERE chat_room_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pg.Pool.Query(ctx, query, id.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("chat rooms list is empty")
		}
		return nil, err
	}
	defer rows.Close()

	var result entity.ChatList
	for rows.Next() {
		var r entity.Chat
		var createdAt time.Time
		err := rows.Scan(&r.ID, &r.ChatRoomID, &r.Request, &r.Response, &createdAt)
		if err != nil {
			return nil, err
		}

		r.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		result.Chats = append(result.Chats, r)
	}

	return &result, nil
}

func (r *ChatRepo) Check(ctx context.Context, userID, chatRoomID string) error {

	fmt.Println("chatRoomID", chatRoomID, "userID", userID)
	var role string
	err := r.pg.Pool.QueryRow(ctx, `
		SELECT role FROM users WHERE id = $1 AND deleted_at = 0
	`, userID).Scan(&role)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	var requestLimit, chatLimit int
	err = r.pg.Pool.QueryRow(ctx, `
		SELECT request_limit, chat_limit FROM restrictions WHERE type = $1
	`, role).Scan(&requestLimit, &chatLimit)
	if err != nil {
		return fmt.Errorf("failed to get restrictions: %w", err)
	}

	var todayRequestCount int
	err = r.pg.Pool.QueryRow(ctx, `
		SELECT COUNT(c.id)
		FROM chat c
		JOIN chat_rooms cr ON cr.id = c.chat_room_id
		WHERE cr.user_id = $1 AND c.created_at::date = CURRENT_DATE AND c.deleted_at = 0
	`, userID).Scan(&todayRequestCount)
	if err != nil {
		return fmt.Errorf("failed to count today's requests: %w", err)
	}

	if todayRequestCount >= requestLimit {
		return errors.New("kunlik request limiti tugadi")
	}

	var chatRequestCount int
	err = r.pg.Pool.QueryRow(ctx, `
		SELECT COUNT(id) FROM chat 
		WHERE chat_room_id = $1 AND deleted_at = 0
	`, chatRoomID).Scan(&chatRequestCount)
	if err != nil {
		return fmt.Errorf("failed to count chat room requests: %w", err)
	}

	if chatRequestCount >= chatLimit {
		return errors.New("bu chatdagi savollar limiti tugagan, yangi chat yarating")
	}

	return nil
}

func (r *ChatRepo) DeleteChatRoom(ctx context.Context, id *entity.ById) error {
	query := `DELETE from chat_rooms WHERE id = $1`
	_, err := r.pg.Pool.Exec(ctx, query, id.Id)
	if err != nil {
		return err
	}
	return nil
}
