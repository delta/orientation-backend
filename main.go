package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delta/orientation-backend/config"
)

func main() {

	config.InitConfig()

	port := config.Config("PORT")
	addr := fmt.Sprintf(":%s",port)

	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorln("Error occured", r)
		}
	}()

	handler := func (w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, "work in progress")
	}

	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(addr, nil))
}
