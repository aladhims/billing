// Package billing provides a simple billing engine for managing loans.
//
// It offers functionality to create loans, make payments, check outstanding
// balances, determine if a loan is delinquent, and manage multiple loans
// through a central Engine.
//
// The main components of this package are:
//
// Loan: Represents an individual loan with its properties and methods.
// It encapsulates all loan-related operations and data.
//
// Engine: Manages multiple loans and provides thread-safe operations
// for creating, retrieving, and manipulating loans.
//
// Key features:
// - Create loans with custom configurations
// - Make payments on loans
// - Check outstanding balances
// - Determine if a loan is delinquent
// - Retrieve billing schedules
// - Get loan statuses
//
// Usage:
//
//	 // Create a new billing engine
//	 engine := billing.NewEngine()
//
//	 // Create a new loan
//	 loan, err := engine.CreateLoan(
//	     billing.WithLoanID("loan1"),
//	     billing.WithLoanConfig(billing.Config{
//	         Principal:    1000000,
//	         InterestRate: 0.10,
//	         TotalWeeks:   50,
//	     }),
//	 )
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//
//	 // Make a payment
//	 err = engine.MakePayment("loan1", loan.GetWeeklyPayment())
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//
//	 // Check outstanding balance
//	 outstanding, err := engine.GetOutstanding("loan1")
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 fmt.Printf("Outstanding balance: %.2f\n", outstanding)
//
//	 // Check if loan is delinquent
//	 isDelinquent, err := engine.IsDelinquent("loan1")
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 fmt.Printf("Is loan delinquent: %v\n", isDelinquent)
//
//	 // Get billing schedule
//	 schedule, err := engine.GetBillingSchedule("loan1")
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 fmt.Printf("Billing schedule: %v\n", schedule)
//
//	 // Get loan status
//	 status, err := engine.GetLoanStatus("loan1")
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 fmt.Printf("Loan status: %v\n", status)
package billing
