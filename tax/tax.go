package tax

const PERSONAL_TAX_DEDUCTION = 60_000

type Tax struct {
}

func New() Tax {
	return Tax{}
}

func (t Tax) calculate(totalIncome float64) float64 {
	incomeAfterPersonalTaxDeduction := totalIncome - PERSONAL_TAX_DEDUCTION

	tax := 0.0

	if incomeAfterPersonalTaxDeduction <= 150_000 {
		tax = 0
	}

	return tax
}
