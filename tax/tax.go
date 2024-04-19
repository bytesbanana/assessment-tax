package tax

const PERSONAL_TAX_DEDUCTION = 60_000

type TaxCalculator struct {
}

func New() TaxCalculator {
	return TaxCalculator{}
}

func (t TaxCalculator) calculate(info TaxInformation) float64 {
	income := t.deductedIncome(info)

	if income <= 150_000 {
		return 0
	} else if income <= 500_000 {

		return (income-150_000.0)*0.1 - info.WHT
	} else if income <= 1_000_000 {
		return (income-500_000)*0.15 + 35_000 - info.WHT
	} else if income <= 2_000_000 {
		return (income-1_000_000)*0.2 + 110_000 - info.WHT
	}

	return (income-2_000_000)*0.35 + 310000 - info.WHT
}

func (t TaxCalculator) deductedIncome(info TaxInformation) float64 {
	baseDeduction := info.TotalIncome - PERSONAL_TAX_DEDUCTION

	sumAllawances := 0.0
	for _, allowance := range info.Allowances {
		sumAllawances += allowance.Amount
	}

	return baseDeduction - sumAllawances
}
