package routes

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiKeyResponse struct {
	Key string `json:"key"`
}

type ApiKeyRequest struct {
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewApiKey(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
	bytes, err := io.ReadAll(r.Body)	

	if err != nil {
		return err
	}

	user := ApiKeyRequest{}

	if err := json.Unmarshal(bytes, &user); err != nil {
		return err
	}

	iotUser := IotUser{}

	err = db.Where("username = ?", user.Username).First(&iotUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(200)
			w.Write([]byte("Your credentials are incorrect!"))

			return nil
		}

		return err
	}

	hash := sha256.Sum256([]byte(user.Password))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	if iotUser.PasswordHash != encoded {
		w.WriteHeader(200)
		w.Write([]byte("Your credentials are incorrect!"))

		return nil
	} 

	apiKeyBytes := make([]byte, 32); rand.Read(apiKeyBytes)

	key := fmt.Sprintf("iot_%s", apiKeyBytes)

	apiKeyHash := sha256.Sum256([]byte(key))

	encodedApiKey := base64.StdEncoding.EncodeToString(apiKeyHash[:])

	apiKey := ApiKey{
		Id: uint(uuid.New().ID()),
		Name: user.Name,
		Hash: encodedApiKey,
		Username: iotUser.Username,
	}

	err = db.Create(&apiKey).Error

	if err != nil {
		return err
	}

	iotUser.ApiKeys = append(iotUser.ApiKeys, apiKey.Name)

	err = db.Save(&iotUser).Error

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte(key))

	return nil
}
