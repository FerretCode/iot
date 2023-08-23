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
	Teams pq.StringArray `json:"teams" gorm:"type:text[]"`
}

func Create(w http.ResponseWriter, r *http.Request, db gorm.DB) error {
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

	for _, v := range teamRequest.InitialUsers {
		user := IotUser{} 

		err := db.Where("username = ?", v).First(&user).Error 

		if err != nil {
			return err
		}

		user.Teams = append(user.Teams, team.Name)

		err = db.Where("username = ?", v).Save(&user).Error

		if err != nil {
			return err
		}
	}

	w.WriteHeader(200)
	w.Write([]byte("The team has been successfully created."))
	
	return nil
}

