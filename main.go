package main

import (
	"encoding/json"
	"log"
	"mux-mongo-api/configs"
	"mux-mongo-api/routes"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	//run database
	configs.ConnectDB()

	//attach the routes
	routes.PurchaseRoute(router)

	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		json.NewEncoder(rw).Encode(map[string]string{"data": "Hello, This is my Multi Currency Purchase Transaction Application with Golang & MongoDB"})
	}).Methods("GET")

	log.Fatal(http.ListenAndServe(":6000", router))

}
