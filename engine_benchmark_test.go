package billing

import (
	"fmt"
	"testing"
)

func BenchmarkEngine_CreateLoan(b *testing.B) {
	engine := NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CreateLoan(WithLoanID(fmt.Sprintf("loan%d", i)))
	}
}

func BenchmarkEngine_GetLoan(b *testing.B) {
	engine := NewEngine()
	_, _ = engine.CreateLoan(WithLoanID("loan1"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetLoan("loan1")
	}
}

func BenchmarkEngine_MakePayment(b *testing.B) {
	engine := NewEngine()
	loan, _ := engine.CreateLoan(WithLoanID("loan1"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.MakePayment("loan1", loan.GetWeeklyPayment())
	}
}
