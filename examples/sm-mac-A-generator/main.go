package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

// Order represents a single order
type Order struct {
	OrderID  string `json:"orderId"`
	Customer string `json:"customer"`
	Amount   int    `json:"amount"`
}

func main() {
	orders := generateOrders(1000_000)

	// Write to file
	err := saveToFile("orders-2.json", orders)
	if err != nil {
		fmt.Printf("Error saving to file: %v\n", err)
		return
	}

	fmt.Println("Successfully generated 1000 orders and saved to orders.json")
	printSummary(orders)
}

func generateOrders(n int) []Order {
	firstNames := []string{
		"Emma", "James", "Sophia", "William", "Olivia", "Liam", "Isabella", "Noah", "Mia", "Ethan",
		"Charlotte", "Benjamin", "Amelia", "Lucas", "Harper", "Alexander", "Evelyn", "Henry", "Abigail",
		"Daniel", "Emily", "Michael", "Elizabeth", "Matthew", "Sofia", "Joseph", "Avery", "David",
		"Ella", "Samuel", "Madison", "John", "Scarlett", "Gabriel", "Victoria", "Anthony", "Aria",
		"Andrew", "Grace", "Joshua", "Chloe", "Christopher", "Camila", "Kevin", "Penelope", "Brian",
		"Riley", "Thomas", "Layla", "Robert", "Mila", "Steven", "Nora", "Richard", "Hannah", "Charles",
		"Zoe", "Thomas", "Lily", "Donald", "Stella", "Mark", "Aurora", "Paul", "Savannah", "George",
		"Anna", "Kenneth", "Lucy", "Steven", "Caroline", "Edward", "Natalie", "Brian", "Genesis",
		"Ronald", "Kennedy", "Timothy", "Samantha", "Jason", "Maya", "Jeffrey", "Naomi", "Ryan",
		"Elena", "Jacob", "Ruby", "Gary", "Ivy", "Nicholas", "Eliana", "Eric", "Madelyn", "Stephen",
		"Clara", "Larry", "Vivian", "Justin", "Josephine", "Scott", "Delilah", "Brandon", "Luna",
	}

	orders := make([]Order, n)

	for i := 0; i < n; i++ {
		orderID := fmt.Sprintf("ORD-%04d", i+1)
		customer := firstNames[rand.Intn(len(firstNames))]
		amount := rand.Intn(991) + 10

		orders[i] = Order{
			OrderID:  orderID,
			Customer: customer,
			Amount:   amount,
		}
	}

	return orders
}

func saveToFile(filename string, orders []Order) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(orders)
}

func printSummary(orders []Order) {
	if len(orders) == 0 {
		return
	}

	total := 0
	minAmount := orders[0].Amount
	maxAmount := orders[0].Amount
	minOrder := orders[0]
	maxOrder := orders[0]
	customerCount := make(map[string]int)

	for _, order := range orders {
		total += order.Amount
		customerCount[order.Customer]++

		if order.Amount < minAmount {
			minAmount = order.Amount
			minOrder = order
		}
		if order.Amount > maxAmount {
			maxAmount = order.Amount
			maxOrder = order
		}
	}

	fmt.Println("\n--- Summary Statistics ---")
	fmt.Printf("Total Orders: %d\n", len(orders))
	fmt.Printf("Total Amount: %d\n", total)
	fmt.Printf("Average Amount: %.2f\n", float64(total)/float64(len(orders)))
	fmt.Printf("Minimum Order: %s - %s: %d\n", minOrder.OrderID, minOrder.Customer, minOrder.Amount)
	fmt.Printf("Maximum Order: %s - %s: %d\n", maxOrder.OrderID, maxOrder.Customer, maxOrder.Amount)
	fmt.Printf("Unique Customers: %d\n", len(customerCount))
}
