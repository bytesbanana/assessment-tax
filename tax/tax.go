package tax

import "math"

var progressiveTax = []struct {
	minIncome float64
	maxIncome float64
	rate      float64
}{
	{
		minIncome: 0,
		maxIncome: 150_000.0,
		rate:      0.0,
	},
	{
		minIncome: 150_000.0,
		maxIncome: 500_000.0,
		rate:      0.1,
	},
	{
		minIncome: 500_000,
		maxIncome: 1_000_000.0,
		rate:      0.15,
	},
	{
		minIncome: 1_000_000.0,
		maxIncome: 2_000_000.0,
		rate:      0.20,
	},
	{
		minIncome: 2_000_000,
		maxIncome: math.MaxFloat64,
		rate:      0.35,
	},
}

const (
	MAX_TAX_LEVEL_0        = 150_000.0
	MAX_TAX_LEVEL_1        = 500_000.0
	MAX_TAX_LEVEL_2        = 1_000_000.0
	MAX_TAX_LEVEL_3        = 2_000_000.0
	MAX_TAX_LEVEL_1_AMOUNT = 35_000
	MAX_TAX_LEVEL_2_AMOUNT = 75_000
	MAX_TAX_LEVEL_3_AMOUNT = 200_000
	MAX_DONATE_DEDUCTION   = 100_000
)

type TaxCalculator struct {
	personalDededucation float64
	maxKReceiptDeduction float64
}

func NewTaxCalculator(personalDeductation float64, maxKReceiptDeduction float64) TaxCalculator {
	return TaxCalculator{
		personalDededucation: personalDeductation,
		maxKReceiptDeduction: maxKReceiptDeduction,
	}
}

type CalculateTaxDetails struct {
	tax       float64
	taxRefund float64
	taxLevel  []TaxLevel
}

func NewTaxDetails() CalculateTaxDetails {
	return CalculateTaxDetails{
		tax:       0,
		taxRefund: 0,
		taxLevel: []TaxLevel{
			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   0.0,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0.0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0.0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0.0,
			},
		},
	}
}

func (t TaxCalculator) calculate(info TaxInformation) CalculateTaxDetails {
	income := t.calDeductedIncome(info)

	details := NewTaxDetails()

	for level, proTax := range progressiveTax {
		if income > proTax.minIncome {
			max := (proTax.maxIncome - proTax.minIncome) * proTax.rate
			tax := math.Min((income-proTax.minIncome)*proTax.rate, max)
			details.tax += tax
			details.taxLevel[level].Tax = tax
		}
	}

	details.taxRefund = t.calTaxRefund(details.tax, info.WHT)
	details.tax = math.Max(details.tax-info.WHT, 0)

	return details
}

func (t TaxCalculator) calDeductedIncome(info TaxInformation) float64 {
	baseDeduction := info.TotalIncome - t.personalDededucation

	kReceiptSum := info.sumAllowanceByType(ACCEPT_ALLOWANCE_TYPES["k-receipt"])
	donationSum := info.sumAllowanceByType(ACCEPT_ALLOWANCE_TYPES["donation"])

	donationDeduction := math.Min(donationSum, MAX_DONATE_DEDUCTION)
	kRecieptDeduction := math.Min(kReceiptSum, t.maxKReceiptDeduction)

	return baseDeduction - donationDeduction - kRecieptDeduction
}

func (t TaxCalculator) calTaxRefund(tax float64, wht float64) float64 {
	if wht <= tax {
		return 0
	}
	return math.Max(wht-tax, 0)
}
