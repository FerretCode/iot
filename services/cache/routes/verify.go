package routes

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ferretcode/iot/services/user/routes"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("the user was not found in the cache")

type VerifyRequest struct {
	ApiKey string `json:"api_key"`
}

func Verify(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	verifyRequest := VerifyRequest{}

	if err := json.Unmarshal(bytes, &verifyRequest); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(verifyRequest.ApiKey))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	key := routes.ApiKey{}

	err = db.Where("hash = ?", encoded).First(&key).Error	

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
		user := routes.IotUser{}

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

	iotUser := routes.IotUser{}

	if err := json.Unmarshal([]byte(result), &iotUser); err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte(result))

	return nil
}
