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
	var mongodbHealth string
	router := mux.NewRouter()

	//run database
	if mongoClient := configs.ConnectDB(); mongoClient != nil {
		mongodbHealth = "success"
	}

	//attach the routes
	routes.PurchaseRoute(router)

	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		json.NewEncoder(rw).Encode(map[string]string{"data": "Hello, This is my Multi Currency Purchase Transaction Application with Golang & MongoDB"})
	}).Methods("GET")

	router.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		type healthStatus struct {
			MongoDB     string `json:"mongodb"`
			PurchaseAPI string `json:"purchaseAPI"`
		}
		//var data map[string]interface{}
		healthResponse := healthStatus{MongoDB: mongodbHealth, PurchaseAPI: "success"}
		//json.Unmarshal([]byte(healthResponse), &data)
		json.NewEncoder(rw).Encode(map[string]interface{}{"health": healthResponse})
	})

	log.Fatal(http.ListenAndServe(":6000", router))

}
