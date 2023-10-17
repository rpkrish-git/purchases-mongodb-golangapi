package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mux-mongo-api/configs"
	"mux-mongo-api/models"
	"mux-mongo-api/responses"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type currencyExchange struct {
	Currency     string
	Rate         string
	RecordedDate string
}

type requestBody struct {
	search inputCurrency
}

type inputCurrency struct {
	CountryCurrency string `json:"currency"`
}

var purchaseCollection *mongo.Collection = configs.GetCollections(configs.DB, "purchases")
var validate = validator.New()

func CreateATransaction() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var transaction models.Transaction
		defer cancel()

		//validate the request body
		if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			response := responses.PurchaseResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&transaction); validationErr != nil {
			rw.WriteHeader(http.StatusBadRequest)
			response := responses.PurchaseResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		newTransaction := models.Transaction{
			Id:                  primitive.NewObjectID(),
			Description:         transaction.Description,
			TransactionDateTime: transaction.TransactionDateTime,
			PurchaseAmount:      transaction.PurchaseAmount,
		}
		result, err := purchaseCollection.InsertOne(ctx, newTransaction)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		rw.WriteHeader(http.StatusCreated)
		response := responses.PurchaseResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}}
		json.NewEncoder(rw).Encode(response)
	}
}

func GetATransaction() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		params := mux.Vars(r)
		var body requestBody
		json.NewDecoder(r.Body).Decode(&body.search)
		transactionId := params["transactionId"]
		var transaction models.Transaction
		defer cancel()

		exchRate, err := getCurrencyExchangeRate(body.search.CountryCurrency)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}
		objId, _ := primitive.ObjectIDFromHex(transactionId)

		err = purchaseCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&transaction)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		convRate, _ := strconv.ParseFloat(strings.TrimSpace(exchRate.Rate), 64)
		convPrice, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", transaction.PurchaseAmount*convRate), 64)
		convTrans := models.ConvertedTransaction{
			Id:                      transaction.Id.String(),
			Description:             transaction.Description,
			OriginalPriceInUSDollar: transaction.PurchaseAmount,
			TransactionDateTime:     transaction.TransactionDateTime,
			ConvertedCurrencyDesc:   exchRate.Currency,
			ExchangeRate:            exchRate.Rate,
			ConvertedPrice:          convPrice,
		}
		rw.WriteHeader(http.StatusOK)
		response := responses.PurchaseResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": convTrans}}
		json.NewEncoder(rw).Encode(response)

	}
}

func EditATransaction() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
		params := mux.Vars(r)
		transactionId := params["transactionId"]
		var transaction models.Transaction
		defer cancle()

		objId, _ := primitive.ObjectIDFromHex(transactionId)

		// validate the request body
		if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		//use the validator library to validate required fields
		if validatationErr := validate.Struct(&transaction); validatationErr != nil {
			rw.WriteHeader(http.StatusBadRequest)
			response := responses.PurchaseResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validatationErr.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		update := bson.M{"description": transaction.Description, "transactionDateTime": transaction.TransactionDateTime, "purchaseAmount": transaction.PurchaseAmount}

		result, err := purchaseCollection.UpdateOne(ctx, bson.M{"id": objId}, bson.M{"$set": update})
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		//get updated user details
		var updatedTransaction models.Transaction
		if result.MatchedCount == 1 {
			err := purchaseCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedTransaction)

			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{}}
				json.NewEncoder(rw).Encode(response)
				return
			}
		}

		rw.WriteHeader(http.StatusOK)
		response := responses.PurchaseResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": updatedTransaction}}
		json.NewEncoder(rw).Encode(response)

	}
}

func DeleteATransaction() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		params := mux.Vars(r)
		transactionId := params["transactionId"]
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(transactionId)

		result, err := purchaseCollection.DeleteOne(ctx, bson.M{"id": objId})

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		if result.DeletedCount < 1 {
			rw.WriteHeader(http.StatusNotFound)
			response := responses.PurchaseResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": "User with specified ID not found"}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		rw.WriteHeader(http.StatusOK)
		response := responses.PurchaseResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "User successfuly deleted"}}
		json.NewEncoder(rw).Encode(response)

	}
}

func GetAllPurchases() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var transactions []models.Transaction
		defer cancel()

		results, err := purchaseCollection.Find(ctx, bson.M{})

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
			json.NewEncoder(rw).Encode(response)
			return
		}

		//reading from the db in the optmial way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleTransaction models.Transaction
			if err := results.Decode(&singleTransaction); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				response := responses.PurchaseResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}}
				json.NewEncoder(rw).Encode(response)
				return
			}
			transactions = append(transactions, singleTransaction)
		}

		rw.WriteHeader(http.StatusOK)
		response := responses.PurchaseResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": transactions}}
		json.NewEncoder(rw).Encode(response)
	}
}

func getCurrencyExchangeRate(countryCurrency string) (currencyExchange, error) {
	var curExch currencyExchange
	fmt.Println("1. Performing Http Get of Currence Exchange Rate for given currency...")
	currExchUrl := fmt.Sprintf("https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?sort=-record_date&format=json&page[number]=1&page[size]=1&fields=country_currency_desc,exchange_rate,record_date&filter=country_currency_desc:in:(%s),record_date:gte:2019-02-01,record_date:lt:2019-07-30", strings.ToUpper(countryCurrency))
	resp, err := http.Get(currExchUrl)
	if err != nil {
		log.Fatalln(err)
		return curExch, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return curExch, errors.New("Error in retrieving the currenct exchange rate.. Please check")
	}
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Convert response body to string
	//bodyString := string(bodyBytes)
	//fmt.Println("API Response as String:\n" + bodyString)

	type currencyData interface{}

	// Convert response body to Todo struct
	type currencyExchangeResponse struct {
		Data []currencyData `json:"data"`
	}

	var currencyConverter currencyExchangeResponse
	json.Unmarshal(bodyBytes, &currencyConverter)
	if len(currencyConverter.Data) == 0 {
		return curExch, errors.New("Purchase cannot be converted to the target currency.. Please check")
	}
	if converterMap, ok := (currencyConverter.Data[0]).(map[string]interface{}); ok {

		// m is a map[string]interface{}
		for k, v := range converterMap {
			switch k {
			case "country_currency_desc":
				curExch.Currency = v.(string)
			case "exchange_rate":
				curExch.Rate = v.(string)
			case "record_date":
				curExch.RecordedDate = v.(string)
			}
		}

	}

	fmt.Printf("API Response of Currency Exchange: %+v\n", curExch)
	return curExch, nil
}
