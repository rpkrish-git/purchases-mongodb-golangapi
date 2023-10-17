package routes

import (
	"mux-mongo-api/controllers"

	"github.com/gorilla/mux"
)

func PurchaseRoute(router *mux.Router) {
	//All routes related to users comes here
	router.HandleFunc("/purchase", controllers.CreateATransaction()).Methods("POST")
	router.HandleFunc("/purchases/{transactionId}", controllers.GetATransaction()).Methods("POST")
	router.HandleFunc("/purchases/{transactionId}", controllers.EditATransaction()).Methods("PUT")
	router.HandleFunc("/purchases/{transactionId}", controllers.DeleteATransaction()).Methods("DELETE")
	router.HandleFunc("/purchases", controllers.GetAllPurchases()).Methods("GET")
	//router.HandleFunc("/health-check", controllers.HealthCheck()).Methods("GET")
}
