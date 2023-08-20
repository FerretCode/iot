package routes

import (
	"net/http"
	"io"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"encoding/json"

	"gorm.io/gorm"
)

type DeleteApiKeyRequest struct {
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func DeleteApiKey(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	deleteApiKeyRequest := DeleteApiKeyRequest{}

	if err := json.Unmarshal(bytes, &deleteApiKeyRequest); err != nil {
		return err
	}

	iotUser := IotUser{}

	err = db.First(&iotUser, deleteApiKeyRequest.Username).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(200)
			w.Write([]byte("Your credentials are incorrect!"))

			return nil
		}

		return err
	}

	hash := sha256.Sum256([]byte(deleteApiKeyRequest.Password))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	if iotUser.PasswordHash != encoded {
		w.WriteHeader(200)
		w.Write([]byte("Your credentials are incorrect!"))

		return nil
	}

	index := 0

	for i, v := range iotUser.ApiKeys {
		if v.Name == deleteApiKeyRequest.Name { index = i }
	} 

	iotUser.ApiKeys = append(
		iotUser.ApiKeys[:index],
		iotUser.ApiKeys[index + 1:]...,
	)

	err = db.Save(&iotUser).Error

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The API key was successfully deleted."))

	return nil
}
