package placement

import "fmt"

// This file defines the handlers that mixed.go references
// This creates the scenario where mixed.go has low cohesion (all external refs)

// HandleUser works with User from mixed.go
func HandleUser(u *User) string {
	return fmt.Sprintf("User: %d", u.ID)
}

// HandleProduct works with Product from mixed.go
func HandleProduct(p *Product) string {
	return fmt.Sprintf("Product: %d", p.ID)
}

// HandleOrder works with Order from mixed.go
func HandleOrder(o *Order) string {
	return fmt.Sprintf("Order: %d", o.ID)
}

// ProcessAll uses all handlers
func ProcessAll() {
	u := &User{ID: 1, Name: "Alice"}
	p := &Product{ID: 1, Title: "Widget", Price: 9.99}
	o := &Order{ID: 1, UserID: u.ID, ProductID: p.ID}
	
	fmt.Println(HandleUser(u))
	fmt.Println(HandleProduct(p))
	fmt.Println(HandleOrder(o))
}
