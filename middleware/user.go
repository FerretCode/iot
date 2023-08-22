package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ferretcode/iot/services/user/routes"
	"gorm.io/gorm"
)

func CheckAPIKey(db gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("Authorization")

			proxyHost := os.Getenv("IOT_CACHE_PROXY_SERVICE_HOST")	
			proxyPort := os.Getenv("IOT_CACHE_PROXY_SERVICE_PORT")	

			body := map[string]string{
				"api_key": apiKey,
			}

			stringified, err := json.Marshal(&body)

			if err != nil {
				HandleError(w, r, err)

				return 
			}

			req, err := http.NewRequest(
				"POST",
				fmt.Sprintf("http://%s:%s/validate", proxyHost, proxyPort),
				bytes.NewReader(stringified),
			)

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

