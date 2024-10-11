package billing

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LoanStatus represents the current status of a loan
type LoanStatus int

// Loan statuses
const (
	Active LoanStatus = iota
	Delinquent
	Closed
)

// Loan-related durations
const (
	DaysPerWeek          = 7
	HoursPerDay          = 24
	DelinquencyThreshold = 2 * DaysPerWeek * HoursPerDay * time.Hour
)

const (
	// DefaultPrincipal is the default loan principal amount in IDR
	DefaultPrincipal = 5_000_000 // 5 million IDR

	// DefaultInterestRate is the default annual interest rate as a decimal
	DefaultInterestRate = 0.10 // 10% per annum

	// DefaultLoanDurationWeeks is the default loan duration in weeks
	DefaultLoanDurationWeeks = 50
)

// Config holds the configuration for loan creation
type Config struct {
	Principal    float64
	InterestRate float64
	TotalWeeks   int
}

// DefaultConfig provides default values for loan configuration
var DefaultConfig = Config{
	Principal:    float64(DefaultPrincipal),
	InterestRate: DefaultInterestRate,
	TotalWeeks:   DefaultLoanDurationWeeks,
}

// Payment represents a single payment made towards a loan
type Payment struct {
	Amount float64
	Date   time.Time
}

// Loan represents a loan with its properties and methods
type Loan struct {
	ID              string
	Principal       float64
	InterestRate    float64
	TotalWeeks      int
	WeeklyPayment   float64
	StartDate       time.Time
	Payments        []Payment
	OutstandingDebt float64
	Status          LoanStatus
}

// LoanOption defines a function type for loan options
type LoanOption func(*Loan)

// WithLoanID sets a custom ID for the loan
func WithLoanID(id string) LoanOption {
	return func(l *Loan) {
		l.ID = id
	}
}

// WithLoanConfig sets a custom configuration for the loan
func WithLoanConfig(config Config) LoanOption {
	return func(l *Loan) {
		l.Principal = config.Principal
		l.InterestRate = config.InterestRate
		l.TotalWeeks = config.TotalWeeks
		
		totalInterest := config.Principal * config.InterestRate
		totalAmount := config.Principal + totalInterest
		l.WeeklyPayment = totalAmount / float64(config.TotalWeeks)
		l.OutstandingDebt = totalAmount
	}
}

// NewLoan creates a new loan with the given options
func NewLoan(options ...LoanOption) *Loan {
	loan := &Loan{
		ID:              uuid.New().String(),
		Principal:       DefaultConfig.Principal,
		InterestRate:    DefaultConfig.InterestRate,
		TotalWeeks:      DefaultConfig.TotalWeeks,
		StartDate:       time.Now(),
		Status:          Active,
	}


	totalInterest := loan.Principal * loan.InterestRate
	totalAmount := loan.Principal + totalInterest
	loan.WeeklyPayment = totalAmount / float64(loan.TotalWeeks)
	loan.OutstandingDebt = totalAmount

	for _, option := range options {
		option(loan)
	}

	return loan
}

// GetOutstanding returns the current outstanding debt of the loan
func (l *Loan) GetOutstanding() float64 {
	return l.OutstandingDebt
}

// IsDelinquent checks if the loan is delinquent
func (l *Loan) IsDelinquent() bool {
	if len(l.Payments) < 2 {
		return false
	}

	lastTwoPayments := l.Payments[len(l.Payments)-2:]
	expectedLastPaymentDate := l.StartDate.Add(time.Duration(len(l.Payments)-1) * DaysPerWeek * HoursPerDay * time.Hour)

	return time.Since(expectedLastPaymentDate) > DelinquencyThreshold &&
		lastTwoPayments[0].Amount == 0 && lastTwoPayments[1].Amount == 0
}

// MakePayment records a payment for the loan
func (l *Loan) MakePayment(amount float64) error {
	currentWeek := int(time.Since(l.StartDate).Hours() / (DaysPerWeek * HoursPerDay))
	expectedPayments := currentWeek + 1 // +1 because payments start from week 0
	actualPayments := len(l.Payments)
	missedPayments := expectedPayments - actualPayments

	if missedPayments > 0 {
		expectedAmount := float64(missedPayments) * l.WeeklyPayment
		if amount < expectedAmount {
			return fmt.Errorf("payment amount must be at least %.2f for %d missed payments", expectedAmount, missedPayments)
		}
	} else if amount != l.WeeklyPayment {
		return errors.New("payment amount must be equal to the weekly payment")
	}

	if l.OutstandingDebt <= 0 {
		return errors.New("loan is already fully paid")
	}

	l.Payments = append(l.Payments, Payment{Amount: amount, Date: time.Now()})
	l.OutstandingDebt -= amount

	if l.OutstandingDebt <= 0 {
		l.Status = Closed
	} else if l.IsDelinquent() {
		l.Status = Delinquent
	} else {
		l.Status = Active
	}

	return nil
}

// GetBillingSchedule returns the weekly payment schedule for the loan
func (l *Loan) GetBillingSchedule() []float64 {
	schedule := make([]float64, l.TotalWeeks)
	for i := range schedule {
		schedule[i] = l.WeeklyPayment
	}
	return schedule
}