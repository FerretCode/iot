package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lib/pq"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type IotUser struct {
	gorm.Model
	Username    string    `json:"username" gorm:"primaryKey"`
	Id          uint      `json:"id"`
	PasswordHash string   `json:"password_hash"`
	ApiKeys     pq.StringArray `json:"api_keys" gorm:"type:text[]"`
	DataPoints  pq.StringArray  `json:"data_points" gorm:"type:text[]"`
	Teams       pq.StringArray `json:"teams" gorm:"type:text[]"`
}

type ApiKey struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name" gorm:"primaryKey"`
	Hash     string `json:"hash"`
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

	db.AutoMigrate(&IotUser{}, &ApiKey{})

	err = db.Where("username = ?", userRequest.Username).First(&iotUser).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("failed to get")

			return err	
		}
	}

	if iotUser.Username != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("this user already exists"))

		return nil
	}

	iotUser.Id = uint(uuid.New().ID())
	iotUser.Username = userRequest.Username
	iotUser.PasswordHash = encoded
	iotUser.ApiKeys = make([]string, 0)
	iotUser.DataPoints = make([]string, 0)
	iotUser.Teams = make([]string, 0)

	err = db.Create(&iotUser).Error

	if err != nil {
		fmt.Println("creation error")

		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("your account has been created successfully"))

	return nil
}
