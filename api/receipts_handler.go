package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"net/http"
	"receipt-processor-challenge/internal/receipt"
)

// ReceiptApi - Receipt api
type ReceiptApi struct {
	db *buntdb.DB
}

// NewReceiptApi - default instance of ReceiptApi
func NewReceiptApi(db *buntdb.DB) *ReceiptApi {
	return &ReceiptApi{
		db: db,
	}
}

// InitializeRoutes defines the routes for the receipts
func (api *ReceiptApi) InitializeRoutes(mux *mux.Router) {
	mux.HandleFunc("/receipts/process", api.receiptProcessor)
	mux.HandleFunc("/receipts/{id}/points", api.getPointsByID)
}

// receiptProcessor REST endpoint /receipts/process
// takes in a receipt object to calculate points and store in db
func (api *ReceiptApi) receiptProcessor(w http.ResponseWriter, r *http.Request) {
	if methodTypeNotAllowed(w, r, "POST") {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var currentReceipt receipt.Receipt
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&currentReceipt)

	if err != nil {
		http.Error(w, "Invalid Json in Body", http.StatusBadRequest)
		return
	}

	rs := receipt.NewBuntDBOrderRepository(api.db)

	createReceipt, err := rs.CreateReceipt(currentReceipt)
	if err != nil {
		http.Error(w, err.Error(), errorCodeAssigner(err))
		return
	}

	receiptUUID := map[string]string{"id": createReceipt}

	marshal, err := json.Marshal(receiptUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getPointsByID REST endpoint /receipts/{id}/points
// takes in a dynamic string: id
// return int: points - amount of points for the string
func (api *ReceiptApi) getPointsByID(w http.ResponseWriter, r *http.Request) {
	if methodTypeNotAllowed(w, r, "GET") {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	receiptID := vars["id"]

	rs := receipt.NewBuntDBOrderRepository(api.db)

	points, err := rs.GetPointsByID(receiptID)

	if err != nil {
		http.Error(w, err.Error(), errorCodeAssigner(err))
		return
	}
	pointsMap := map[string]int{"points": points}
	marshal, err := json.Marshal(pointsMap)

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshal)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func methodTypeNotAllowed(w http.ResponseWriter, r *http.Request, methodType string) bool {
	if r.Method != methodType {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return true
	}
	return false
}

func errorCodeAssigner(err error) int {
	switch err.Error() {
	case "duplicate receipt", "exist":
		return 409
	case "not found", "Not found":
		return 404

	default:
		return 500
	}
}
