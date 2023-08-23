package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Team struct {
	Id uint `json:"id"`
	Name string `json:"name" gorm:"primaryKey"`
	Users pq.StringArray `json:"users" gorm:"type:text[]"`
	DataPoints pq.StringArray `json:"data_points" gorm:"type:text[]"`
	Secure bool `json:"secure"`
}

type TeamRequest struct {
	Name string `json:"name"`
	InitialUsers []string `json:"initial_users"`
	Secure bool `json:"secure"`
}

type IotUser struct {
	Username string `json:"username"`
}

func Create(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
	fmt.Println(r.Context().Value("user"))

	user := r.Context().Value("user").(IotUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	teamRequest := TeamRequest{}

	if err := json.Unmarshal(bytes, &teamRequest); err != nil {
		return err
	}

	team := Team{}

	err = db.Where("name = ?", teamRequest.Name).First(&team).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	if team.Name != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("this team already exists"))

		return nil
	}	

	teamRequest.InitialUsers = append(teamRequest.InitialUsers, user.Username)

	team.Id = uint(uuid.New().ID())
	team.Name = teamRequest.Name
	team.Users = append(team.Users, teamRequest.InitialUsers...)
	team.DataPoints = make([]string, 0)
	team.Secure = team.Secure

	err = db.Create(&team).Error

	if err != nil {
		return err
	}
	
	return nil
}

