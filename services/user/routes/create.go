package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"gorm.io/gorm"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type IotUser struct {
	Username string `json:"username"`
	PasswordHash string `json:"password_hash"`
	ApiKeys []ApiKey `json:"api_keys"`
	DataPoints []string `json:"data_points"`
	Teams []string `json:"teams"`
}

type ApiKey struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

func Create(w http.ResponseWriter, r *http.Request, db gorm.DB) error {	
	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	userRequest := UserRequest{}

	if err := json.Unmarshal(bytes, &userRequest); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(userRequest.Password))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	iotUser := IotUser{}

	db.First(&iotUser, userRequest.Username)

	if iotUser.Username != "" {
		w.WriteHeader(200)
		w.Write([]byte("you have been logged in successfully"))

		return nil
	}

	iotUser.Username = userRequest.Username
	iotUser.PasswordHash = encoded
	iotUser.ApiKeys = make([]ApiKey, 0)
	iotUser.DataPoints = make([]string, 0)
	iotUser.Teams = make([]string, 0)

	db.Create(iotUser)

	w.WriteHeader(200)
	w.Write([]byte("your account has been created successfully"))

	return nil
}
