package nearclone

import "fmt"

// near_clone.go contains near duplicates (Type 3 clones)
// These have similar structure but with some differences (above 80% similarity threshold)

func ProcessOrderA(orderID string, items []string) error {
	if orderID == "" {
		return ErrInvalidOrder
	}
	if len(items) == 0 {
		return ErrEmptyOrder
	}
	inventory := GetInventory()
	for _, item := range items {
		if !inventory.Has(item) {
			return ErrItemNotFound
		}
	}
	order := CreateOrder(orderID, items)
	return SaveOrder(order)
}

func ProcessOrderB(orderID string, items []string) error {
	if orderID == "" {
		return ErrInvalidOrder
	}
	if len(items) == 0 {
		return ErrEmptyOrder
	}
	inventory := GetInventory()
	for _, item := range items {
		if !inventory.Has(item) {
			return ErrItemNotFound
		}
		inventory.Reserve(item) // Additional line - creates near duplicate
	}
	order := CreateOrder(orderID, items)
	return SaveOrder(order)
}

// FormatReportA and FormatReportB are similar but not identical
func FormatReportA(data map[string]int) string {
	result := "Report:\n"
	for key, value := range data {
		result += key + ": " + string(value) + "\n"
	}
	result += "End of report"
	return result
}

func FormatReportB(data map[string]int) string {
	output := "Report:\n"
	total := 0
	for key, value := range data {
		output += key + ": " + string(value) + "\n"
		total += value // Additional calculation
	}
	output += "Total: " + string(total) + "\n" // Additional line
	output += "End of report"
	return output
}

var (
	ErrInvalidOrder  = fmt.Errorf("invalid order")
	ErrEmptyOrder    = fmt.Errorf("empty order")
	ErrItemNotFound  = fmt.Errorf("item not found")
)

type Inventory struct{}

func GetInventory() *Inventory {
	return &Inventory{}
}

func (i *Inventory) Has(item string) bool {
	return true
}

func (i *Inventory) Reserve(item string) {}

type Order struct{}

func CreateOrder(id string, items []string) *Order {
	return &Order{}
}

func SaveOrder(order *Order) error {
	return nil
}
