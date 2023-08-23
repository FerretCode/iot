package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type InvalidateRequest struct {
	Hash string `json:"hash"`
}

func Invalidate(w http.ResponseWriter, r *http.Request, db gorm.DB, ctx context.Context) error {
	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	invalidateRequest := InvalidateRequest{}

	if err := json.Unmarshal(bytes, &invalidateRequest); err != nil {
		return err
	}

	ip := os.Getenv("IOT_CACHE_SERVICE_HOST")
	port := os.Getenv("IOT_CACHE_SERVICE_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", ip, port),
		Password: "",
		DB: 0,
	})

	_, err = rdb.Del(ctx, invalidateRequest.Hash).Result()

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The API key was successfully invalidated."))

	return nil
}
