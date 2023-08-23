package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ferretcode/iot/services/user/routes"
	"gorm.io/gorm"
)

func CheckAPIKey(db gorm.DB, proxyHost string, proxyPort string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("Authorization")

			req, err := http.NewRequest(
				"POST",
				fmt.Sprintf("http://%s:%s/verify", proxyHost, proxyPort),
				nil,
			)

			req.Header.Set("Authorization", apiKey)

			if err != nil {
				HandleError(w, r, err)

				return
			}

			res, err := http.DefaultClient.Do(req)

			if err != nil {
				HandleError(w, r, err)

				return
			}

			bytes, err := io.ReadAll(res.Body)

			if err != nil {
				HandleError(w, r, err)

				return
			}

			fmt.Println(string(bytes))

			user := routes.IotUser{}

			if err := json.Unmarshal(bytes, &user); err != nil {
				HandleError(w, r, err)

				return
			}

			if user.Username == "" {
				w.WriteHeader(http.StatusUnauthorized)	
				w.Write([]byte("The API key is invalid!"))

				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println(err)

	http.Error(w, "There was an error verifying you.", http.StatusInternalServerError)

	return
}

