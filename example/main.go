package main

import (
	"fmt"
	"log"

	"github.com/aladhims/billing"
)

func main() {
	// Create a new billing engine
	engine := billing.NewEngine()

	// Create a new loan with custom configuration
	loanID := "loan001"
	loan, err := engine.CreateLoan(
		billing.WithLoanID(loanID),
		billing.WithLoanConfig(billing.Config{
			Principal:    1000000,
			InterestRate: 0.10, // 10% per annum
			TotalWeeks:   50,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create loan: %v", err)
	}

	fmt.Printf("Loan created: ID=%s, Weekly Payment=%.2f\n", loan.GetID(), loan.GetWeeklyPayment())

	// Get the initial outstanding balance
	outstanding, err := engine.GetOutstanding(loanID)
	if err != nil {
		log.Fatalf("Failed to get outstanding balance: %v", err)
	}
	fmt.Printf("Initial outstanding balance: %.2f\n", outstanding)

	// Get and print the billing schedule
	schedule, err := engine.GetBillingSchedule(loanID)
	if err != nil {
		log.Fatalf("Failed to get billing schedule: %v", err)
	}
	fmt.Println("Billing Schedule:")
	for i, payment := range schedule {
		fmt.Printf("Week %d: %.2f\n", i+1, payment)
	}

	// Make some payments
	paymentAmount := loan.GetWeeklyPayment()
	for i := 0; i < 3; i++ {
		err = engine.MakePayment(loanID, paymentAmount)
		if err != nil {
			log.Fatalf("Failed to make payment: %v", err)
		}
		fmt.Printf("Made payment of %.2f\n", paymentAmount)

		// Check new outstanding balance
		outstanding, err = engine.GetOutstanding(loanID)
		if err != nil {
			log.Fatalf("Failed to get outstanding balance: %v", err)
		}
		fmt.Printf("New outstanding balance: %.2f\n", outstanding)
	}

	// Check if the loan is delinquent (it shouldn't be)
	isDelinquent, err := engine.IsDelinquent(loanID)
	if err != nil {
		log.Fatalf("Failed to check delinquency: %v", err)
	}
	fmt.Printf("Is loan delinquent? %v\n", isDelinquent)

	// Simulate missed payments to make the loan delinquent
	// fmt.Println("\nSimulating missed payments...") -- can't do it without explicitly change the unexported start date field
	// time.Sleep(3 * billing.DaysPerWeek * billing.HoursPerDay * time.Hour)

	// Check delinquency again
	// isDelinquent, _ = engine.IsDelinquent(loanID)
	// fmt.Printf("Is loan delinquent after missed payments? %v\n", isDelinquent)

	// Try to make a payment less than the required amount
	err = engine.MakePayment(loanID, paymentAmount)
	fmt.Printf("Attempting to make a payment of %.2f\n", paymentAmount)
	if err != nil {
		fmt.Printf("Payment failed as expected: %v\n", err)
	}

	// Make a payment to cover missed payments
	requiredPayment := 4 * paymentAmount // 4 weeks of payments (3 missed + 1 current)
	err = engine.MakePayment(loanID, requiredPayment)
	if err != nil {
		fmt.Printf("Failed to make payment: %v\n", err)
	} else {
		fmt.Printf("Made payment of %.2f to cover missed payments\n", requiredPayment)
	}

	// Check final status
	isDelinquent, _ = engine.IsDelinquent(loanID)
	fmt.Printf("Is loan still delinquent? %v\n", isDelinquent)

	outstanding, _ = engine.GetOutstanding(loanID)
	fmt.Printf("Final outstanding balance: %.2f\n", outstanding)
}
