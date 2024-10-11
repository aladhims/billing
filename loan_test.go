package billing

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLoan(t *testing.T) {
	tests := []struct {
		name     string
		options  []LoanOption
		expected Config
	}{
		{
			name:     "Default loan",
			options:  []LoanOption{},
			expected: DefaultConfig,
		},
		{
			name: "Custom ID",
			options: []LoanOption{
				WithLoanID("custom-id"),
			},
			expected: DefaultConfig,
		},
		{
			name: "Custom config",
			options: []LoanOption{
				WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.05,
					TotalWeeks:   25,
				}),
			},
			expected: Config{
				Principal:    1000000,
				InterestRate: 0.05,
				TotalWeeks:   25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := NewLoan(tt.options...)

			assert.Equal(t, tt.expected.Principal, loan.GetPrincipal())
			assert.Equal(t, tt.expected.InterestRate, loan.GetInterestRate())
			assert.Equal(t, tt.expected.TotalWeeks, loan.GetTotalWeeks())
			assert.Equal(t, Active, loan.GetStatus())

			if tt.name == "Custom ID" {
				assert.Equal(t, "custom-id", loan.GetID())
			} else {
				_, err := uuid.Parse(loan.GetID())
				assert.NoError(t, err, "ID should be a valid UUID")
			}

			assert.InDelta(t, time.Now().Unix(), loan.GetStartDate().Unix(), 1, "StartDate should be close to now")

			expectedTotal := tt.expected.Principal * (1 + tt.expected.InterestRate)
			assert.InDelta(t, expectedTotal, loan.GetOutstanding(), 0.01, "OutstandingDebt should be principal plus interest")
			assert.InDelta(t, expectedTotal/float64(tt.expected.TotalWeeks), loan.GetWeeklyPayment(), 0.01, "WeeklyPayment should be total divided by weeks")
		})
	}
}

func TestLoan_GetOutstanding(t *testing.T) {
	loan := NewLoan(WithLoanConfig(Config{
		Principal:    1000000,
		InterestRate: 0.10,
		TotalWeeks:   50,
	}))

	assert.Equal(t, 1100000.0, loan.GetOutstanding(), "Initial outstanding should be principal plus interest")

	_ = loan.MakePayment(22000)

	assert.Equal(t, 1078000.0, loan.GetOutstanding(), "Outstanding should decrease after payment")
}

func TestLoan_IsDelinquent(t *testing.T) {
	tests := []struct {
		name      string
		setupLoan func() *Loan
		expected  bool
	}{
		{
			name: "No payments, within two weeks",
			setupLoan: func() *Loan {
				loan := NewLoan()
				loan.startDate = time.Now().Add(-13 * 24 * time.Hour)
				return loan
			},
			expected: false,
		},
		{
			name: "No payments, over two weeks",
			setupLoan: func() *Loan {
				loan := NewLoan()
				loan.startDate = time.Now().Add(-15 * 24 * time.Hour)
				return loan
			},
			expected: true,
		},
		{
			name: "Last payment within two weeks",
			setupLoan: func() *Loan {
				loan := NewLoan()
				loan.MakePayment(loan.GetWeeklyPayment())
				loan.payments[0].Date = time.Now().Add(-13 * 24 * time.Hour)
				return loan
			},
			expected: false,
		},
		{
			name: "Last payment over two weeks ago",
			setupLoan: func() *Loan {
				loan := NewLoan()
				loan.MakePayment(loan.GetWeeklyPayment())
				loan.payments[0].Date = time.Now().Add(-15 * 24 * time.Hour)
				return loan
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := tt.setupLoan()
			assert.Equal(t, tt.expected, loan.IsDelinquent())
		})
	}
}

func TestLoan_MakePayment(t *testing.T) {
	tests := []struct {
		name           string
		setupLoan      func() *Loan
		paymentAmount  float64
		expectedError  string
		expectedDebt   float64
		expectedStatus LoanStatus
	}{
		{
			name: "Regular payment - Active",
			setupLoan: func() *Loan {
				return NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
			},
			paymentAmount:  22000,
			expectedError:  "",
			expectedDebt:   1078000,
			expectedStatus: Active,
		},
		{
			name: "Payment too small",
			setupLoan: func() *Loan {
				return NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
			},
			paymentAmount:  21000,
			expectedError:  "payment amount must be at least 22000.00 for 1 missed payments",
			expectedDebt:   1100000,
			expectedStatus: Active,
		},
		{
			name: "Payment after loan is closed",
			setupLoan: func() *Loan {
				loan := NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
				for loan.GetOutstanding() > 0 {
					loan.MakePayment(loan.GetWeeklyPayment())
				}
				return loan
			},
			paymentAmount:  22000,
			expectedError:  "loan is already fully paid",
			expectedDebt:   0,
			expectedStatus: Closed,
		},
		{
			name: "Payment after missed payments - becomes Active",
			setupLoan: func() *Loan {
				loan := NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
				loan.startDate = time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour)
				return loan
			},
			paymentAmount:  88000,
			expectedError:  "",
			expectedDebt:   1012000,
			expectedStatus: Active,
		},
		{
			name: "Payment that closes the loan",
			setupLoan: func() *Loan {
				loan := NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
				loan.outstandingDebt = 22000
				return loan
			},
			paymentAmount:  22000,
			expectedError:  "",
			expectedDebt:   0,
			expectedStatus: Closed,
		},
		{
			name: "Payment after becoming delinquent",
			setupLoan: func() *Loan {
				loan := NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
				loan.startDate = time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour)
				loan.payments = []Payment{
					{Amount: 0, Date: time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour)},
					{Amount: 0, Date: time.Now().Add(-2 * DaysPerWeek * HoursPerDay * time.Hour)},
				}
				return loan
			},
			paymentAmount:  44000,
			expectedError:  "",
			expectedDebt:   1056000,
			expectedStatus: Active,
		},
		{
			name: "Insufficient payment for delinquent loan",
			setupLoan: func() *Loan {
				loan := NewLoan(WithLoanConfig(Config{
					Principal:    1000000,
					InterestRate: 0.10,
					TotalWeeks:   50,
				}))
				loan.startDate = time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour)
				loan.payments = []Payment{
					{Amount: 0, Date: time.Now().Add(-3 * DaysPerWeek * HoursPerDay * time.Hour)},
					{Amount: 0, Date: time.Now().Add(-2 * DaysPerWeek * HoursPerDay * time.Hour)},
				}
				return loan
			},
			paymentAmount:  22000,
			expectedError:  "payment amount must be at least 44000.00 for 2 missed payments",
			expectedDebt:   1100000,
			expectedStatus: Active,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := tt.setupLoan()
			err := loan.MakePayment(tt.paymentAmount)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.InDelta(t, tt.expectedDebt, loan.GetOutstanding(), 0.01)
			assert.Equal(t, tt.expectedStatus, loan.GetStatus(), "Loan status should match expected status")
		})
	}
}

func TestLoan_GetBillingSchedule(t *testing.T) {
	loan := NewLoan(WithLoanConfig(Config{
		Principal:    1000000,
		InterestRate: 0.10,
		TotalWeeks:   50,
	}))

	schedule := loan.GetBillingSchedule()

	assert.Len(t, schedule, 50, "Schedule should have 50 weeks")
	for _, payment := range schedule {
		assert.InDelta(t, 22000, payment, 0.01, "Each payment should be 22000")
	}
}
