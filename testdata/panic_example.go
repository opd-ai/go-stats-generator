package testlib

import (
	"fmt"
	"log"
)

// GoodFunction returns an error instead of panicking
func GoodFunction(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}

// BadFunctionWithPanic uses panic in library code
func BadFunctionWithPanic(data string) {
	if data == "" {
		panic("empty data")
	}
}

// BadFunctionWithLogFatal uses log.Fatal in library code
func BadFunctionWithLogFatal(data string) {
	if data == "" {
		log.Fatal("empty data")
	}
}

// init is allowed to panic for configuration errors
func init() {
	if false {
		panic("configuration error")
	}
}
