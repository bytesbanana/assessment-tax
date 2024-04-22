package tax

type TaxInformation struct {
	TotalIncome float64     `json:"totalIncome"`
	WHT         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

func (t *TaxInformation) sumAllowanceByType(allowanceType string) float64 {
	kReceiptSum := 0.0
	for _, allowance := range t.Allowances {
		if allowance.AllowanceType == allowanceType {
			kReceiptSum += allowance.Amount
		}
	}
	return kReceiptSum
}
