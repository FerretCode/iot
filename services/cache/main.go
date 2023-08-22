package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ferretcode/iot/services/cache/routes"
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

	r.Route("/api/cache", func(r chi.Router) {
		r.Post("/verify", func(w http.ResponseWriter, r *http.Request) {
			err := routes.Verify(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error verifying the user!", http.StatusInternalServerError)
			}
		})
	})

	http.ListenAndServe(":3000", r)
}
