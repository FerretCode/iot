package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	iotMiddleware "github.com/ferretcode/iot/middleware"
	"github.com/ferretcode/iot/services/teams/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	scrapartydb "github.com/scraparty/scraparty-db"
)

func main() {
	db, err := scrapartydb.Connect()

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(iotMiddleware.CheckAPIKey(
		db,
		os.Getenv("IOT_CACHE_SERVICE_HOST"),
		os.Getenv("IOT_CACHE_SERVICE_PORT"),
	))

	r.Route("/api/teams", func(r chi.Router) {
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			err := routes.Create(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error creating the team!", http.StatusInternalServerError)
			}
		})
	})

	http.ListenAndServe(":3000", r)
}
