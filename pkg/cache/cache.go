package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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
