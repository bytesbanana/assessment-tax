package tax

import "math"

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

	if income > MAX_TAX_LEVEL_0 {
		tax := math.Min((income-MAX_TAX_LEVEL_0)*0.1, MAX_TAX_LEVEL_1_AMOUNT)
		details.tax += tax
		details.taxLevel[1].Tax = tax
	}

	if income > MAX_TAX_LEVEL_1 {
		tax := math.Min((income-MAX_TAX_LEVEL_1)*0.15, MAX_TAX_LEVEL_2_AMOUNT)
		details.tax += tax
		details.taxLevel[2].Tax = tax
	}

	if income > MAX_TAX_LEVEL_2 {
		tax := math.Min((income-MAX_TAX_LEVEL_2)*0.2, MAX_TAX_LEVEL_3_AMOUNT)
		details.tax += tax
		details.taxLevel[3].Tax = tax
	}

	if income > MAX_TAX_LEVEL_3 {
		tax := (income - MAX_TAX_LEVEL_3) * 0.35
		details.tax += tax
		details.taxLevel[4].Tax = tax
	}

	details.taxRefund = t.calTaxRefund(details.tax, info.WHT)
	details.tax = math.Max(details.tax-info.WHT, 0)

	return details
}

func (t TaxCalculator) calDeductedIncome(info TaxInformation) float64 {
	baseDeduction := info.TotalIncome - t.personalDededucation

	kReceiptSum := sumAllowanceByType(info.Allowances, ACCEPT_ALLOWANCE_TYPES["k-receipt"])
	donationSum := sumAllowanceByType(info.Allowances, ACCEPT_ALLOWANCE_TYPES["donation"])

	donationDeduction := math.Min(donationSum, MAX_DONATE_DEDUCTION)
	kRecieptDeduction := math.Min(kReceiptSum, t.maxKReceiptDeduction)

	return baseDeduction - donationDeduction - kRecieptDeduction
}

func sumAllowanceByType(allowances []Allowance, allowanceType string) float64 {
	kReceiptSum := 0.0
	for _, allowance := range allowances {
		if allowance.AllowanceType == allowanceType {
			kReceiptSum += allowance.Amount
		}
	}
	return kReceiptSum
}

func (t TaxCalculator) calTaxRefund(tax float64, wht float64) float64 {
	if wht <= tax {
		return 0
	}
	return math.Max(wht-tax, 0)
}
