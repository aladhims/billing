package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T, *Engine)
	}{
		{"CreateLoan", testCreateLoan},
		{"GetLoan", testGetLoan},
		{"GetOutstanding", testGetOutstanding},
		{"IsDelinquent", testIsDelinquent},
		{"MakePayment", testMakePayment},
		{"GetBillingSchedule", testGetBillingSchedule},
		{"GetLoanStatus", testGetLoanStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			tt.testFunc(t, engine)
		})
	}
}

func testCreateLoan(t *testing.T, engine *Engine) {
	tests := []struct {
		name        string
		loanID      string
		expectError bool
	}{
		{"Create new loan", "loan1", false},
		{"Create duplicate loan", "loan1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan, err := engine.CreateLoan(WithLoanID(tt.loanID))
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loan)
				assert.Equal(t, tt.loanID, loan.GetID())
			}
		})
	}
}

func testGetLoan(t *testing.T, engine *Engine) {
	_, _ = engine.CreateLoan(WithLoanID("loan1"))

	tests := []struct {
		name        string
		loanID      string
		expectError bool
	}{
		{"Get existing loan", "loan1", false},
		{"Get non-existent loan", "non-existent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan, err := engine.GetLoan(tt.loanID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loan)
				assert.Equal(t, tt.loanID, loan.GetID())
			}
		})
	}
}

func testGetOutstanding(t *testing.T, engine *Engine) {
	_, _ = engine.CreateLoan(WithLoanID("loan1"), WithLoanConfig(Config{
		Principal:    1000000,
		InterestRate: 0.10,
		TotalWeeks:   50,
	}))

	tests := []struct {
		name           string
		loanID         string
		expectError    bool
		expectedAmount float64
	}{
		{"Get outstanding for existing loan", "loan1", false, 1100000.0},
		{"Get outstanding for non-existent loan", "non-existent", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outstanding, err := engine.GetOutstanding(tt.loanID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, outstanding)
			}
		})
	}
}

func testIsDelinquent(t *testing.T, engine *Engine) {
	loan, _ := engine.CreateLoan(WithLoanID("loan1"))

	tests := []struct {
		name             string
		loanID           string
		setupFunc        func()
		expectError      bool
		expectDelinquent bool
	}{
		{
			name:             "New loan not delinquent",
			loanID:           "loan1",
			setupFunc:        func() {},
			expectError:      false,
			expectDelinquent: false,
		},
		{
			name:             "Loan becomes delinquent",
			loanID:           "loan1",
			setupFunc:        func() { loan.startDate = time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour) },
			expectError:      false,
			expectDelinquent: true,
		},
		{
			name:             "Non-existent loan",
			loanID:           "non-existent",
			setupFunc:        func() {},
			expectError:      true,
			expectDelinquent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			isDelinquent, err := engine.IsDelinquent(tt.loanID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectDelinquent, isDelinquent)
			}
		})
	}
}

func testMakePayment(t *testing.T, engine *Engine) {
	loan, _ := engine.CreateLoan(WithLoanID("loan1"), WithLoanConfig(Config{
		Principal:    1000000,
		InterestRate: 0.10,
		TotalWeeks:   50,
	}))

	tests := []struct {
		name        string
		loanID      string
		amount      float64
		expectError bool
	}{
		{"Make valid payment", "loan1", loan.GetWeeklyPayment(), false},
		{"Make payment for non-existent loan", "non-existent", 1000, true},
		{"Make invalid payment amount", "loan1", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.MakePayment(tt.loanID, tt.amount)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func testGetBillingSchedule(t *testing.T, engine *Engine) {
	_, _ = engine.CreateLoan(WithLoanID("loan1"), WithLoanConfig(Config{
		Principal:    1000000,
		InterestRate: 0.10,
		TotalWeeks:   50,
	}))

	tests := []struct {
		name        string
		loanID      string
		expectError bool
		expectedLen int
	}{
		{"Get schedule for existing loan", "loan1", false, 50},
		{"Get schedule for non-existent loan", "non-existent", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := engine.GetBillingSchedule(tt.loanID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, schedule, tt.expectedLen)
			}
		})
	}
}

func testGetLoanStatus(t *testing.T, engine *Engine) {
	_, _ = engine.CreateLoan(WithLoanID("loan1"))

	tests := []struct {
		name           string
		loanID         string
		expectError    bool
		expectedStatus LoanStatus
	}{
		{"Get status for existing loan", "loan1", false, Active},
		{"Get status for non-existent loan", "non-existent", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := engine.GetLoanStatus(tt.loanID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}
