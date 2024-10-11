package billing

import (
	"errors"
	"sync"
)

// Engine manages loans
type Engine struct {
	loans map[string]*Loan
	mutex sync.RWMutex
}

// NewEngine creates a new loan engine
func NewEngine() *Engine {
	return &Engine{
		loans: make(map[string]*Loan),
	}
}

// CreateLoan creates a new loan and stores it in the engine
func (e *Engine) CreateLoan(options ...LoanOption) (*Loan, error) {
	loan := NewLoan(options...)

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.loans[loan.GetID()]; exists {
		return nil, errors.New("loan with this ID already exists")
	}

	e.loans[loan.GetID()] = loan
	return loan, nil
}

// GetLoan retrieves a loan by its ID
func (e *Engine) GetLoan(id string) (*Loan, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	loan, exists := e.loans[id]
	if !exists {
		return nil, errors.New("loan not found")
	}
	return loan, nil
}

// GetOutstanding gets the outstanding amount for a specific loan
func (e *Engine) GetOutstanding(id string) (float64, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	loan, exists := e.loans[id]
	if !exists {
		return 0, errors.New("loan not found")
	}

	return loan.GetOutstanding(), nil
}

// IsDelinquent checks if a specific loan is delinquent
func (e *Engine) IsDelinquent(id string) (bool, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	loan, exists := e.loans[id]
	if !exists {
		return false, errors.New("loan not found")
	}

	return loan.IsDelinquent(), nil
}

// MakePayment makes a payment for a specific loan
func (e *Engine) MakePayment(id string, amount float64) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	loan, exists := e.loans[id]
	if !exists {
		return errors.New("loan not found")
	}

	return loan.MakePayment(amount)
}

// GetBillingSchedule returns the billing schedule for a specific loan
func (e *Engine) GetBillingSchedule(id string) ([]float64, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	loan, exists := e.loans[id]
	if !exists {
		return nil, errors.New("loan not found")
	}

	return loan.GetBillingSchedule(), nil
}

// GetLoanStatus returns the status of a specific loan
func (e *Engine) GetLoanStatus(id string) (LoanStatus, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	loan, exists := e.loans[id]
	if !exists {
		return 0, errors.New("loan not found")
	}

	return loan.GetStatus(), nil
}
