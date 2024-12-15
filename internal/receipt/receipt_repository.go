package receipt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/buntdb"
	"math"
	"strings"
	"time"
	"unicode"
)

type ReceiptRepository interface {
	// CreateReceipt Creates a receipt and returns a (UUID, error)
	CreateReceipt(receipt Receipt) (string, error)
	// GetReceiptByID Gets receipt with UUID and returns (Receipt, error)
	GetReceiptByID(id string) (Receipt, error)
	// GetPointsByID Get points by receipt UUID and returns (points, error)
	GetPointsByID(id string) (string, error)
}

type BuntDBReceiptRepository struct {
	db *buntdb.DB
}

func NewBuntDBOrderRepository(db *buntdb.DB) *BuntDBReceiptRepository {
	return &BuntDBReceiptRepository{db: db}
}

// CreateReceipt - creates a receipt from a Receipt
// returns (Receipt.id, error)
func (repo *BuntDBReceiptRepository) CreateReceipt(receipt Receipt) (string, error) {

	uuidMaxCollisions := 3

	receipt.Points = pointGenerator(receipt)
	var err error

	//truncates receipt.retailer and concatenates other receipt attributes to form a uniqueKey
	uniqueKey := fmt.Sprintf("%s+%s+%s+%f", strings.TrimSpace(receipt.Retailer), receipt.PurchaseDate, receipt.PurchaseTime, receipt.Total)

	// checks for duplicate receipts database
	err = repo.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(uniqueKey)
		return err
	})

	if err == nil {
		return "", errors.New("duplicate receipt")
	}

	err = nil

	var receiptID uuid.UUID

	// attempts to generate uuid for Receipt with a maximum of 3 tries: collisions are astronomically low
	for i := 0; i < uuidMaxCollisions; i++ {
		receiptID, err = uuid.NewUUID()
		if err != nil {
			continue
		}

		err = repo.db.View(func(tx *buntdb.Tx) error {
			_, err := tx.Get("receipt:" + receiptID.String())

			return err
		})

		if err != nil {
			break
		}
	}

	// three attempts to create UUID failed
	if err == nil {
		return "", errors.New("could not create UUID for receipt")
	}

	receipt.Id = receiptID.String()
	receiptMarshal, err := json.Marshal(receipt)

	if err != nil {
		return "", err
	}

	// creates db entry for Receipt with uuid being the primary key
	err = repo.db.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set("receipt:"+receiptID.String(), string(receiptMarshal), nil)
		return err
	})

	// creates composite key with the concatenated string generated to detect duplicate Receipts
	if err == nil {
		err = repo.db.Update(func(tx *buntdb.Tx) error {
			_, _, err = tx.Set(uniqueKey, receiptID.String(), nil)
			return err
		})
		// returns the receipt uuid, error
		return receiptID.String(), err
	}

	return "", err
}

// GetPointsByID - takes a receipt number and returns points for that receipt
func (repo *BuntDBReceiptRepository) GetPointsByID(id string) (int, error) {

	var receipt *Receipt
	var stringValue string
	var err error

	err = repo.db.View(func(tx *buntdb.Tx) error {
		stringValue, err = tx.Get("receipt:" + id)
		return err
	})

	if err == nil {
		err = json.Unmarshal([]byte(stringValue), &receipt)
		if err == nil {
			return receipt.Points, err
		}
	}
	return 0, err
}

// pointGenerator - calculates reward points based on the following criteria:
// 1. Retailer Name:
//   - +1 point for every alphanumeric character in the retailer name.
//
// 2. Total Amount:
//   - +50 points if the total is a round dollar amount with no cents.
//   - +25 points if the total is a multiple of 0.25.
//
// 3. Receipt Items:
//   - +5 points for every two items on the receipt.
//   - If the trimmed length of an item's description is a multiple of 3:
//   - Multiply the item price by 0.2, round up to the nearest integer, and add the result as points.
//
// 4. Purchase Date & Time:
//   - +6 points if the day of the purchase date is odd.
//   - +10 points if the purchase time is between 2:00 PM and 4:00 PM.
func pointGenerator(receipt Receipt) int {

	var (
		points                     int
		retailer                   = receipt.Retailer
		items                      = receipt.Items
		total                      = receipt.Total
		receiptPurchaseDateTime, _ = time.Parse("2006-01-02 15:04", receipt.PurchaseDate+" "+receipt.PurchaseTime)
	)

	for _, chars := range retailer {
		if unicode.IsDigit(chars) || unicode.IsLetter(chars) {
			points++
		}
	}

	for _, item := range items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			points += int(math.Ceil(item.Price * .2))
			//points += int(currentItemPrice.Mul(decimal.NewFromFloat(.2)).Ceil().IntPart())
		}
	}

	if math.Mod(total, 1.0) == 0 {
		points += 50

	}

	if math.Mod(total, .25) == 0 {
		points += 25
	}

	points += (len(items) / 2) * 5

	if receiptPurchaseDateTime.Day()%2 != 0 {
		points += 6
	}

	if isPurchaseAfterTwoBeforeFour(receiptPurchaseDateTime) {
		points += 10
	}

	return points
}

// isPurchaseAfterTwoBeforeFour - returns true if time is between 14:00 PM and 16:00 PM
func isPurchaseAfterTwoBeforeFour(time time.Time) bool {
	return (time.Hour() > 14 || (time.Hour() == 14 && time.Minute() > 1)) && time.Hour() < 16
}
