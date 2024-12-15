package main

import (
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"net/http"
	"receipt-processor-challenge/api"
)

func main() {
	// in memory database
	db, err := buntdb.Open(":memory")

	if err != nil {
		panic(err)
	}
	// mux route handler
	handler := mux.NewRouter() // uses gorilla mux for handling dynamic routing

	// Receipt API /receipts
	receiptApi := api.NewReceiptApi(db)

	// inits receipt endpoints
	receiptApi.InitializeRoutes(handler)

	// establishes http server on port 8080
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
