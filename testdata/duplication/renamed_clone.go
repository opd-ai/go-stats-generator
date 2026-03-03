package renamedclone

import "fmt"

// renamed_clone.go contains renamed duplicates (Type 2 clones)
// These have identical structure but different variable/parameter names

// CalculateTotalPrice computes the sum of prices for given items with discount.
func CalculateTotalPrice(items []string, prices map[string]float64) float64 {
	total := 0.0
	for _, item := range items {
		if price, exists := prices[item]; exists {
			total += price
		}
	}
	if total > 100.0 {
		total *= 0.9
	}
	return total
}

// ComputeFinalCost calculates the total cost of products with discount.
func ComputeFinalCost(products []string, costs map[string]float64) float64 {
	sum := 0.0
	for _, product := range products {
		if cost, found := costs[product]; found {
			sum += cost
		}
	}
	if sum > 100.0 {
		sum *= 0.9
	}
	return sum
}

// HandleRequestA and HandleRequestB have identical structure with different names
func HandleRequestA(requestID string, payload map[string]string) error {
	if requestID == "" {
		return ErrInvalidRequest
	}
	validator := NewValidator()
	if err := validator.Validate(payload); err != nil {
		return err
	}
	processor := NewProcessor()
	return processor.Process(payload)
}

// HandleRequestB processes a request by validating and processing its payload.
func HandleRequestB(reqIdentifier string, data map[string]string) error {
	if reqIdentifier == "" {
		return ErrInvalidRequest
	}
	val := NewValidator()
	if e := val.Validate(data); e != nil {
		return e
	}
	proc := NewProcessor()
	return proc.Process(data)
}

var ErrInvalidRequest = fmt.Errorf("invalid request")

// Validator provides payload validation functionality.
type Validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks the provided data map for correctness.
func (v *Validator) Validate(data map[string]string) error {
	return nil
}

// Processor handles request data processing operations.
type Processor struct{}

// NewProcessor creates a new Processor instance.
func NewProcessor() *Processor {
	return &Processor{}
}

// Process executes the processing logic on the provided data map.
func (p *Processor) Process(data map[string]string) error {
	return nil
}
