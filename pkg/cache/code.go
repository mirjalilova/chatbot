package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func SaveVerificationCode(r *redis.Client, ctx context.Context, phoneNumber, code string, ttl time.Duration) error {
	key := fmt.Sprintf("verify:%s:code", phoneNumber)
	return r.Set(ctx, key, code, ttl).Err()
}

func GetVerificationCode(r *redis.Client, ctx context.Context, phoneNumber string) (string, error) {
	key := fmt.Sprintf("verify:%s:code", phoneNumber)
	return r.Get(ctx, key).Result()
}

func DeleteVerificationCode(r *redis.Client, ctx context.Context, phoneNumber string) error {
	key := fmt.Sprintf("verify:%s:code", phoneNumber)
	return r.Del(ctx, key).Err()
}
