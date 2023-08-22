package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ferretcode/iot/services/user/routes"
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

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			err := routes.Create(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error creating the user!", http.StatusInternalServerError)
			}
		})

		r.Post("/new-api-key", func(w http.ResponseWriter, r *http.Request) {
			err := routes.NewApiKey(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error creating the API key!", http.StatusInternalServerError)
			}
		})

		r.Post("/delete-api-key", func(w http.ResponseWriter, r *http.Request) {
			err := routes.DeleteApiKey(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error deleting the API key!", http.StatusInternalServerError)
			}
		})
	})
	
	http.ListenAndServe(":3000", r)
}
