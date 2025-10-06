package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Organization struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	FoundedYear  int    `json:"founded_year"`
	Industry     string `json:"industry"`
	Headquarters struct {
		Address string `json:"address"`
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"headquarters"`
	Contacts struct {
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Website     string `json:"website"`
		SocialMedia struct {
			LinkedIn  string `json:"linkedin"`
			Twitter   string `json:"twitter"`
			Instagram string `json:"instagram"`
			Telegram  string `json:"telegram"`
		} `json:"social_media"`
	} `json:"contacts"`
	KeyPeople []struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Email string `json:"email"`
	} `json:"key_people"`
	Subsidiaries []struct {
		Name     string `json:"name"`
		Industry string `json:"industry"`
		Location string `json:"location"`
	} `json:"subsidiaries"`
	Registration struct {
		TaxID              string `json:"tax_id"`
		RegistrationNumber string `json:"registration_number"`
	} `json:"registration"`
}

func AppendChatOrganization(r *redis.Client, ctx context.Context, chatID string, orgs []Organization) error {
	now := time.Now().Unix()
	orgKey := fmt.Sprintf("chat:%s:organization:%d", chatID, now)
	orgSetKey := fmt.Sprintf("chat:%s:organizations", chatID)

	data, err := json.Marshal(orgs)
	if err != nil {
		return err
	}

	if err := r.Set(ctx, orgKey, data, time.Hour).Err(); err != nil {
		return err
	}

	return r.ZAdd(ctx, orgSetKey, redis.Z{
		Score:  float64(now),
		Member: orgKey,
	}).Err()
}

func GetChatOrganizations(r *redis.Client, ctx context.Context, chatID string, lastN int64) ([]Organization, error) {
	orgSetKey := fmt.Sprintf("chat:%s:organizations", chatID)

	keys, err := r.ZRevRange(ctx, orgSetKey, 0, lastN-1).Result()
	if err != nil {
		return nil, err
	}

	var organizations []Organization
	for _, key := range keys {
		val, err := r.Get(ctx, key).Result()
		if err == redis.Nil {
			_, _ = r.ZRem(ctx, orgSetKey, key).Result()
			continue
		} else if err != nil {
			return nil, err
		}

		var org Organization
		if err := json.Unmarshal([]byte(val), &org); err != nil {
			return nil, err
		}

		organizations = append(organizations, org)
	}

	return organizations, nil
}

func AppendUserQuery(r *redis.Client, ctx context.Context, userID, query string) error {
	now := time.Now().Unix()
	queryKey := fmt.Sprintf("user:%s:chat:%d", userID, now)
	historySetKey := fmt.Sprintf("user:%s:chat_history", userID)

	err := r.Set(ctx, queryKey, query, time.Hour).Err()
	if err != nil {
		return err
	}

	return r.ZAdd(ctx, historySetKey, redis.Z{
		Score:  float64(now),
		Member: queryKey,
	}).Err()
}

func GetUserQueries(r *redis.Client, ctx context.Context, userID string, lastN int64) ([]string, error) {
	historySetKey := fmt.Sprintf("user:%s:chat_history", userID)

	keys, err := r.ZRevRange(ctx, historySetKey, 0, lastN-1).Result()
	if err != nil {
		return nil, err
	}

	var queries []string
	for _, key := range keys {
		val, err := r.Get(ctx, key).Result()
		if err == redis.Nil {
			_, _ = r.ZRem(ctx, historySetKey, key).Result()
			continue
		} else if err != nil {
			return nil, err
		}
		queries = append(queries, val)
	}

	return queries, nil
}

func ClearUserHistory(r *redis.Client, ctx context.Context, userID string) error {
	historySetKey := fmt.Sprintf("user:%s:chat_history", userID)

	keys, err := r.ZRange(ctx, historySetKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		_ = r.Del(ctx, key).Err()
	}

	return r.Del(ctx, historySetKey).Err()
}
