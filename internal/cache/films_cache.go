package cache

import (
	"REST_api_appl/internal/models"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const (
	filmsListKey = "films:list"
	ttl          = 5 * time.Minute
)

type FilmsCache struct {
	rdb *redis.Client
}

func NewFilmsCache(rdb *redis.Client) *FilmsCache {
	return &FilmsCache{rdb: rdb}
}

func (c *FilmsCache) GetList(ctx context.Context) ([]models.Film, error) {
	data, err := c.rdb.Get(ctx, filmsListKey).Bytes()
	if err != nil {
		return nil, err
	}

	var films []models.Film
	if err := json.Unmarshal(data, &films); err != nil {
		return nil, err
	}
	return films, nil
}

func (c *FilmsCache) SetList(ctx context.Context, films []models.Film) error {
	data, err := json.Marshal(films)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, filmsListKey, data, ttl).Err()
}

func (c *FilmsCache) GetListByID(id int, ctx context.Context) (*models.Film, error) {
	data, err := c.rdb.Get(ctx, strconv.Itoa(id)).Result()
	if err != nil {
		return nil, err
	}

	var film models.Film

	err = json.Unmarshal([]byte(data), &film)
	if err != nil {
		return nil, err
	}

	return &film, nil
}

func (c *FilmsCache) Invalidate(ctx context.Context) error {
	return c.rdb.Del(ctx, filmsListKey).Err()
}
