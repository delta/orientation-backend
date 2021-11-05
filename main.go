package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/rs/cors"
)

func main() {

	config.InitConfig()

	port := config.Config("PORT")
	addr := fmt.Sprintf(":%s", port)

	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorln("Error occured", r)
		}
	}()

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000*"},
		AllowedMethods:   []string{"GET", "OPTIONS", "CORS"},
		AllowedHeaders:   []string{"X-Requested-With", "Content-Type", "Authorization"},
		AllowCredentials: true,
		// Debug: true,
	})
	http.HandleFunc("/", handler)
	http.HandleFunc("/auth", auth.Auth)
	http.HandleFunc("/auth/callback", auth.CallBack)
	http.Handle("/auth/logout", c.Handler(http.HandlerFunc(auth.LogOut)))
	http.Handle("/auth/checkAuth", c.Handler(http.HandlerFunc(auth.CheckAuth)))

	log.Fatal(http.ListenAndServe(addr, nil))
}
