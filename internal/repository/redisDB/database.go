package redisDB

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"urleater/dto"
)

type Storage struct {
	redisClient *redis.Client
}

func NewStorage(redisClient *redis.Client) *Storage {
	return &Storage{
		redisClient: redisClient,
	}
}

func (s *Storage) GetShortLinkByLongLink(ctx context.Context, shortLink string) (*dto.Link, error) {
	res, err := s.redisClient.Get(ctx, shortLink).Result()

	if err != nil {
		return nil, fmt.Errorf("error while getting short link by long link from redis %w", err)
	}

	values := strings.Split(res, "::::")

	if len(values) != 3 {
		return nil, fmt.Errorf("error while getting short link by long link from redis %s", res)
	}

	expiresAt, err := time.Parse(time.RFC3339, values[1])

	if err != nil {
		return nil, fmt.Errorf("error while parsing short link expiry time %w", err)
	}

	return &dto.Link{
		ShortUrl:  shortLink,
		ExpiresAt: expiresAt,
		UserEmail: values[2],
		LongUrl:   values[0],
	}, nil
}

func (s *Storage) SaveShortLinkToLongLink(ctx context.Context, link dto.Link) error {
	value := fmt.Sprintf("%s::::%s::::%s", link.LongUrl, link.ExpiresAt.Format(time.RFC3339), link.UserEmail)

	_, err := s.redisClient.Set(ctx, link.ShortUrl, value, 0).Result()

	if err != nil {
		return fmt.Errorf("error while saving short link to redis %w", err)
	}

	return nil
}

func (s *Storage) DeleteLongLinkByShortLink(ctx context.Context, shortLink string) error {
	_, err := s.redisClient.Del(ctx, shortLink).Result()

	if err != nil {
		return fmt.Errorf("error while deleting short link from redis %w", err)
	}

	return nil
}
