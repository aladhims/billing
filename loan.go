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
	TotalWeeks:   DefaultLoanDurationWeeks, // TODO: can consider to improve on duration instead of weekly-based duration
}

// Payment represents a single payment made towards a loan
type Payment struct {
	Amount float64
	Date   time.Time
}

// Loan represents a loan with its properties and methods
type Loan struct {
	id              string
	principal       float64
	interestRate    float64
	totalWeeks      int
	weeklyPayment   float64
	startDate       time.Time
	payments        []Payment
	outstandingDebt float64
	status          LoanStatus
}

// LoanOption defines a function type for loan options
type LoanOption func(*Loan)

// WithLoanID sets a custom ID for the loan
func WithLoanID(id string) LoanOption {
	return func(l *Loan) {
		l.id = id
	}
}

// WithLoanConfig sets a custom configuration for the loan
func WithLoanConfig(config Config) LoanOption {
	return func(l *Loan) {
		l.principal = config.Principal
		l.interestRate = config.InterestRate
		l.totalWeeks = config.TotalWeeks

		totalInterest := config.Principal * config.InterestRate
		totalAmount := config.Principal + totalInterest
		l.weeklyPayment = totalAmount / float64(config.TotalWeeks)
		l.outstandingDebt = totalAmount
	}
}

// NewLoan creates a new loan with the given options
func NewLoan(options ...LoanOption) *Loan {
	loan := &Loan{
		id:           uuid.New().String(),
		principal:    DefaultConfig.Principal,
		interestRate: DefaultConfig.InterestRate,
		totalWeeks:   DefaultConfig.TotalWeeks,
		startDate:    time.Now(),
		status:       Active,
	}

	totalInterest := loan.principal * loan.interestRate
	totalAmount := loan.principal + totalInterest
	loan.weeklyPayment = totalAmount / float64(loan.totalWeeks)
	loan.outstandingDebt = totalAmount

	for _, option := range options {
		option(loan)
	}

	return loan
}

// GetID returns the ID of the loan
func (l *Loan) GetID() string {
	return l.id
}

// GetOutstanding returns the current outstanding debt of the loan
func (l *Loan) GetOutstanding() float64 {
	return l.outstandingDebt
}

// GetPrincipal returns the principal amount of the loan
func (l *Loan) GetPrincipal() float64 {
	return l.principal
}

// GetInterestRate returns the interest rate of the loan
func (l *Loan) GetInterestRate() float64 {
	return l.interestRate
}

// GetTotalWeeks returns the total number of weeks for the loan
func (l *Loan) GetTotalWeeks() int {
	return l.totalWeeks
}

// GetWeeklyPayment returns the weekly payment amount
func (l *Loan) GetWeeklyPayment() float64 {
	return l.weeklyPayment
}

// GetStartDate returns the start date of the loan
func (l *Loan) GetStartDate() time.Time {
	return l.startDate
}

// GetStatus returns the current status of the loan
func (l *Loan) GetStatus() LoanStatus {
	return l.status
}

// GetPayments returns a copy of the payments slice
func (l *Loan) GetPayments() []Payment {
	paymentsCopy := make([]Payment, len(l.payments))
	copy(paymentsCopy, l.payments)
	return paymentsCopy
}

// IsDelinquent checks if the loan is delinquent
func (l *Loan) IsDelinquent() bool {
	if len(l.payments) > 0 {
		lastPaymentDate := l.payments[len(l.payments)-1].Date
		return time.Since(lastPaymentDate) > DelinquencyThreshold
	}

	return time.Since(l.startDate) > DelinquencyThreshold
}

// MakePayment records a payment for the loan
func (l *Loan) MakePayment(amount float64) error {
	currentWeek := int(time.Since(l.startDate).Hours() / (DaysPerWeek * HoursPerDay))
	expectedPayments := currentWeek + 1 // +1 because payments start from week 0
	actualPayments := len(l.payments)
	missedPayments := expectedPayments - actualPayments

	if missedPayments > 0 {
		expectedAmount := float64(missedPayments) * l.weeklyPayment
		if amount < expectedAmount {
			return fmt.Errorf("payment amount must be at least %.2f for %d missed payments", expectedAmount, missedPayments)
		}
	} else if amount != l.weeklyPayment {
		return errors.New("payment amount must be equal to the weekly payment")
	}

	if l.outstandingDebt <= 0 {
		return errors.New("loan is already fully paid")
	}

	l.payments = append(l.payments, Payment{Amount: amount, Date: time.Now()})
	l.outstandingDebt -= amount

	if l.outstandingDebt <= 0 {
		l.status = Closed
	} else if l.IsDelinquent() {
		l.status = Delinquent
	} else {
		l.status = Active
	}

	return nil
}

// GetBillingSchedule returns the weekly payment schedule for the loan
func (l *Loan) GetBillingSchedule() []float64 {
	schedule := make([]float64, l.totalWeeks)
	for i := range schedule {
		schedule[i] = l.weeklyPayment
	}
	return schedule
}
