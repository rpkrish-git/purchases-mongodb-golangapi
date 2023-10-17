package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	Id                  primitive.ObjectID `json:"id,omitempty"`
	Description         string             `json:"description,omitempty" validate:"required"`
	TransactionDateTime time.Time          `json:"transactionDateTime,omitempty" validate:"required"`
	PurchaseAmount      float64            `json:"purchaseAmount,omitempty" validate:"required"`
}

type ConvertedTransaction struct {
	Id                      string    `json:"id,omitempty"`
	Description             string    `json:"description,omitempty" validate:"required"`
	TransactionDateTime     time.Time `json:"transactionDateTime,omitempty" validate:"required"`
	OriginalPriceInUSDollar float64   `json:"originalPriceInUSDollar,omitempty" validate:"required"`
	ConvertedCurrencyDesc   string    `json:"convertedCurrencyDesc"`
	ExchangeRate            string    `json:"exchangeRate"`
	ConvertedPrice          float64   `json:"convertedPrice"`
}
