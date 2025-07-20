package simple

import (
	"fmt"
	"sort"
	"strings"
)

// Calculator provides mathematical operations
type Calculator struct {
	history []Operation
}

// Operation represents a mathematical operation
type Operation struct {
	Type   string
	Args   []float64
	Result float64
}

// NewCalculator creates a new calculator
func NewCalculator() *Calculator {
	return &Calculator{
		history: make([]Operation, 0),
	}
}

// Add adds two numbers
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.recordOperation("add", []float64{a, b}, result)
	return result
}

// Subtract subtracts two numbers
func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.recordOperation("subtract", []float64{a, b}, result)
	return result
}

// Multiply multiplies two numbers
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.recordOperation("multiply", []float64{a, b}, result)
	return result
}

// Divide divides two numbers
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}

	result := a / b
	c.recordOperation("divide", []float64{a, b}, result)
	return result, nil
}

// recordOperation records an operation in history
func (c *Calculator) recordOperation(opType string, args []float64, result float64) {
	op := Operation{
		Type:   opType,
		Args:   args,
		Result: result,
	}
	c.history = append(c.history, op)
}

// GetHistory returns the operation history
func (c *Calculator) GetHistory() []Operation {
	return c.history
}

// ClearHistory clears the operation history
func (c *Calculator) ClearHistory() {
	c.history = c.history[:0]
}

// Statistics calculates statistics from a slice of numbers
func Statistics(numbers []float64) map[string]float64 {
	if len(numbers) == 0 {
		return map[string]float64{}
	}

	// Sort for median calculation
	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)
	sort.Float64s(sorted)

	// Calculate sum
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}

	// Calculate mean
	mean := sum / float64(len(numbers))

	// Calculate median
	var median float64
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	} else {
		median = sorted[len(sorted)/2]
	}

	// Calculate variance
	variance := 0.0
	for _, num := range numbers {
		diff := num - mean
		variance += diff * diff
	}
	variance /= float64(len(numbers))

	return map[string]float64{
		"count":    float64(len(numbers)),
		"sum":      sum,
		"mean":     mean,
		"median":   median,
		"variance": variance,
		"min":      sorted[0],
		"max":      sorted[len(sorted)-1],
	}
}

// VeryComplexFunction demonstrates a function with very high complexity
func VeryComplexFunction(data map[string]interface{}, options map[string]bool) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	result := make(map[string]interface{})

	for key, value := range data {
		if options["process_strings"] {
			if strVal, ok := value.(string); ok {
				if options["uppercase"] {
					result[key] = strings.ToUpper(strVal)
				} else if options["lowercase"] {
					result[key] = strings.ToLower(strVal)
				} else {
					result[key] = strVal
				}
				continue
			}
		}

		if options["process_numbers"] {
			if floatVal, ok := value.(float64); ok {
				if options["square"] {
					result[key] = floatVal * floatVal
				} else if options["sqrt"] {
					if floatVal < 0 {
						return nil, fmt.Errorf("cannot take square root of negative number")
					}
					// Simplified square root calculation
					result[key] = floatVal / 2 // Simplified for demo
				} else {
					result[key] = floatVal
				}
				continue
			}

			if intVal, ok := value.(int); ok {
				floatVal := float64(intVal)
				if options["square"] {
					result[key] = floatVal * floatVal
				} else if options["sqrt"] {
					if floatVal < 0 {
						return nil, fmt.Errorf("cannot take square root of negative number")
					}
					result[key] = floatVal / 2 // Simplified for demo
				} else {
					result[key] = floatVal
				}
				continue
			}
		}

		if options["process_arrays"] {
			if arrVal, ok := value.([]interface{}); ok {
				newArr := make([]interface{}, 0, len(arrVal))
				for _, item := range arrVal {
					if strItem, ok := item.(string); ok && options["process_strings"] {
						if options["uppercase"] {
							newArr = append(newArr, strings.ToUpper(strItem))
						} else if options["lowercase"] {
							newArr = append(newArr, strings.ToLower(strItem))
						} else {
							newArr = append(newArr, strItem)
						}
					} else if numItem, ok := item.(float64); ok && options["process_numbers"] {
						if options["square"] {
							newArr = append(newArr, numItem*numItem)
						} else {
							newArr = append(newArr, numItem)
						}
					} else {
						newArr = append(newArr, item)
					}
				}
				result[key] = newArr
				continue
			}
		}

		// Default case
		result[key] = value
	}

	return result, nil
}
