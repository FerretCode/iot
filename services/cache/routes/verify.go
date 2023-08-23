package routes

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// actually just a wrapper for the real IotUser
type IotUser struct {
	Username string `json:"username"`
	ApiKeys pq.StringArray `json:"api_keys" gorm:"type:text[]"`
	DataPoints pq.StringArray `json:"data_points" gorm:"type:text[]"`
	Teams pq.StringArray `json:"teams" gorm:"type:text[]"`
}

type ApiKey struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name" gorm:"primaryKey"`
	Hash     string `json:"hash"`
}

var ErrNotFound = errors.New("the user was not found in the cache")

func Verify(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
	apiKey := r.Header.Get("Authorization")

	hash := sha256.Sum256([]byte(apiKey))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	key := ApiKey{}

	err := db.Where("hash = ?", encoded).First(&key).Error	

	if err != nil {
		return err
	}

	ctx := context.Background()

	ip := os.Getenv("IOT_CACHE_SERVICE_HOST")
	port := os.Getenv("IOT_CACHE_SERVICE_PORT")
	
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", ip, port),
		Password: "",
		DB: 0,
	})

	result, err := rdb.Get(ctx, key.Username).Result()

	if err == redis.Nil {
		user := IotUser{}

		err = db.Where("username = ?", key.Username).First(&user).Error

		if err != nil {
			return err
		}

		stringified, err := json.Marshal(user)

		if err != nil {
			return err
		}

		err = rdb.Set(ctx, user.Username, string(stringified), 0).Err()

		if err != nil {
			return err
		}

		w.WriteHeader(200)
		w.Write(stringified)

		return nil
	}

	if err != nil {
		return err
	}

	userResponseWrapper := IotUser{}

	if err := json.Unmarshal([]byte(result), &userResponseWrapper); err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte(result))

	return nil
}
