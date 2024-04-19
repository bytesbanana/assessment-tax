package tax

const PERSONAL_TAX_DEDUCTION = 60_000

type TaxCalculator struct {
}

func New() TaxCalculator {
	return TaxCalculator{}
}

func (t TaxCalculator) calculate(totalIncome float64, wht float64) float64 {
	income := t.deductedIncome(totalIncome)

	if income <= 150_000 {
		return 0
	} else if income <= 500_000 {
		return (income-150_000)*0.1 - wht
	} else if income <= 1_000_000 {
		return (income-500_000)*0.15 + 35_000 - wht
	} else if income <= 2_000_000 {
		return (income-1_000_000)*0.2 + 110_000 - wht
	}

	return (income-2_000_000)*0.35 + 310000 - wht
}

func (t TaxCalculator) deductedIncome(totalIncome float64) float64 {
	return totalIncome - PERSONAL_TAX_DEDUCTION
}
