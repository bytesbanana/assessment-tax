package tax

const PERSONAL_TAX_DEDUCTION = 60_000

type TaxCalculator struct {
}

func New() TaxCalculator {
	return TaxCalculator{}
}

func (t TaxCalculator) calculate(totalIncome float64) float64 {
	income := totalIncome - PERSONAL_TAX_DEDUCTION

	if income <= 150_000 {
		return 0
	} else if income <= 500_000 {
		return (income - 150_000) * 0.1
	}

	return 0
}
