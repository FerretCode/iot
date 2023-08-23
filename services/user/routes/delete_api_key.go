package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
)

type DeleteApiKeyRequest struct {
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type InvalidateRequest struct {
	Hash string `json:"hash"`
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

	err = db.Where("username = ?", deleteApiKeyRequest.Username).First(&iotUser).Error

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
	apiKey := ApiKey{}

	for i, v := range iotUser.ApiKeys {
		if v == deleteApiKeyRequest.Name {
			key := ApiKey{}

			err = db.Where("name = ?", v).First(&key).Error

			if err != nil {
				return err
			}

			if apiKey.Name != "" {
				index = i
			}
			
			apiKey = key
		}
	} 

	if apiKey.Name == "" {
		w.WriteHeader(404)
		w.Write([]byte("This API key does not exist!"))

		return nil
	}

	err = db.Delete(&apiKey).Error

	if err != nil {
		return err
	}

	iotUser.ApiKeys = append(
		iotUser.ApiKeys[:index],
		iotUser.ApiKeys[index + 1:]...,
	)

	fmt.Println(iotUser)

	invalidateRequest := InvalidateRequest{
		Hash: encoded,
	}	

	stringified, err := json.Marshal(invalidateRequest)

	if err != nil {
		return err
	}

	host := os.Getenv("IOT_CACHE_SERVICE_HOST")	
	port := os.Getenv("IOT_CACHE_SERVICE_PORT_PROXY")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%s/invalidate", host, port),
		strings.NewReader(string(stringified)),	
	)

	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		http.Error(w, "There was an error deleting the API key.", http.StatusInternalServerError)

		return nil
	}

	err = db.Save(&iotUser).Error

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The API key was successfully deleted."))

	return nil
}
