package redis_client

import (
	"context"
	"log"

	"woahtify-backend/utils"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedis(ctx context.Context) *Redis {
	redisAddr, err := utils.GetEnv("REDIS_ADDR")
	if err != nil {
		log.Print(err)
		return nil
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Could not connect to Redis: %v", err)
		return nil
	}
	log.Println("Connected to Redis")
	return &Redis{
		client: redisClient,
		ctx:    ctx,
	}
}

func (r *Redis) Ping() (string, error) {
	return r.client.Ping(r.ctx).Result()
}
