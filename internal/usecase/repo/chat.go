package repo

import (
	"context"
	"database/sql"
	"encoding/json"
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
	query := `
		INSERT INTO chat (
			chat_room_id, user_request, gemini_request, responce, citation_urls, location, images_url, organizations
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`

	var id string
	err := r.pg.Pool.QueryRow(ctx, query,
		req.ChatRoomID,
		req.UserRequest,
		req.GeminiRequest,
		req.Responce,
		req.CitationURLs,
		req.Location,
		req.ImagesURL,
		req.Organizations,
	).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepo) GetChatRoomByUserId(ctx context.Context, req *entity.GetChatRoomReq) (*entity.ChatRoomList, error) {
	query := `
		SELECT COUNT(id) OVER () AS total_count, id, title, created_at
		FROM chat_rooms
		WHERE user_id = $1 AND deleted_at = 0`

	var args []interface{}
	args = append(args, req.UserId)

	if req.Limit != 0 {
		query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		args = append(args, req.Limit, req.Offset)
	} else {
		query += " ORDER BY created_at DESC"
	}

	rows, err := r.pg.Pool.Query(ctx, query, args...)
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
		SELECT id, chat_room_id, user_request, responce, citation_urls, location, images_url, organizations, created_at
		FROM chat
		WHERE chat_room_id = $1 AND deleted_at = 0
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
		var (
			id         string
			chatRoomID string
			userReq    string
			response   string
			citations  []string
			locations  []byte
			images     []string
			orgs       []byte
			createdAt  time.Time
		)

		err := rows.Scan(&id, &chatRoomID, &userReq, &response, &citations, &locations, &images, &orgs, &createdAt)
		if err != nil {
			return nil, err
		}

		createdStr := createdAt.Format("2006-01-02 15:04:05")

		var locParsed []map[string]float64
		if len(locations) > 0 {
			if err := json.Unmarshal(locations, &locParsed); err != nil {
				return nil, fmt.Errorf("failed to unmarshal locations: %w", err)
			}
		}
		// USER message
		result.Chats = append(result.Chats, entity.ChatResponce{
			ID:         id,
			ChatRoomID: chatRoomID,
			Role:       "user",
			Content: entity.ContentRes{
				Text: userReq,
			},
			CreatedAt: createdStr,
		})

		// ASSISTANT message
		result.Chats = append(result.Chats, entity.ChatResponce{
			ID:         id,
			ChatRoomID: chatRoomID,
			Role:       "assistant",
			Content: entity.ContentRes{
				Text:          response,
				Citations:     citations,
				Location:      locParsed,
				ImagesURL:     images,
				Organizations: json.RawMessage(orgs),
			},
			CreatedAt: createdStr,
		})
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
	query := `UPDATE chat_rooms SET deleted_at = EXTRACT(EPOCH FROM NOW())::bigint WHERE id = $1`
	_, err := r.pg.Pool.Exec(ctx, query, id.Id)
	if err != nil {
		return err
	}
	return nil
}
