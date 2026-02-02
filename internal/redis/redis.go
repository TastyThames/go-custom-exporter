package redis

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

func queryRedisRoleAndPing(addr, password string, timeout time.Duration) (role string, reachable bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return "", false
	}

	info, err := rdb.Info(ctx, "replication").Result()
	if err != nil {
		return "", true
	}

	return parseRoleFromInfo(info), true
}

func parseRoleFromInfo(info string) string {
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if strings.HasPrefix(line, "role:") {
			return strings.TrimPrefix(line, "role:")
		}
	}
	return ""
}
