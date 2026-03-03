package placement

// This file has low cohesion - contains functions that primarily reference
// symbols from handlers.go rather than referencing each other

// User represents a user
type User struct {
	ID   int
	Name string
}

// Product represents a product
type Product struct {
	ID    int
	Title string
	Price float64
}

// Order represents an order
type Order struct {
	ID        int
	UserID    int
	ProductID int
}

// These functions ALL reference external symbols (from handlers.go)
// None of them reference each other, creating low cohesion

// FormatUser uses HandleUser from handlers.go
func FormatUser(u *User) string {
	return HandleUser(u)
}

// FormatProduct uses HandleProduct from handlers.go
func FormatProduct(p *Product) string {
	return HandleProduct(p)
}

// FormatOrder uses HandleOrder from handlers.go
func FormatOrder(o *Order) string {
	return HandleOrder(o)
}

// Process1 uses ProcessAll from handlers.go
func Process1() {
	ProcessAll()
}

// Process2 uses ProcessAll from handlers.go
func Process2() {
	ProcessAll()
}

// Process3 uses ProcessAll from handlers.go
func Process3() {
	ProcessAll()
}

// Process4 uses ProcessAll from handlers.go
func Process4() {
	ProcessAll()
}

// Process5 uses ProcessAll from handlers.go
func Process5() {
	ProcessAll()
}

// Process6 uses ProcessAll from handlers.go
func Process6() {
	ProcessAll()
}
