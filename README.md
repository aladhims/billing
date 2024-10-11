# Billing Engine

A simple Billing Engine for managing loans and payments.

## Features

- Create and manage loans
- Process payments
- Check loan status and delinquency
- Generate billing schedules

## Usage

```go
engine := billing.NewEngine()

loan, err := engine.CreateLoan(
    billing.WithLoanID("loan1"),
    billing.WithLoanConfig(billing.Config{
        Principal:    1000000,
        InterestRate: 0.10,
        TotalWeeks:   50,
    }),
)

err = engine.MakePayment("loan1", 22000)

outstanding, err := engine.GetOutstanding("loan1")

isDelinquent, err := engine.IsDelinquent("loan1")
```
