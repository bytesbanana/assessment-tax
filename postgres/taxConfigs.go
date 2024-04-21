package postgres

import "time"

type TaxConfig struct {
	ID        int       `postgres:"id"`
	Key       string    `postgres:"key"`
	Name      string    `postgres:"name"`
	Value     float64   `postgres:"value"`
	CreatedAt time.Time `postgres:"created_at"`
	CreatedBy string    `postgres:"created_by"`
	UpdatedAt time.Time `postgres:"updated_at"`
	UpdatedBy string    `postgres:"updated_by"`
}

func (p *Postgres) GetTaxConfig(key string) (*TaxConfig, error) {
	var taxConfig *TaxConfig

	row := p.Db.QueryRow("SELECT * FROM tax_config WHERE key = $1", key)

	err := row.Scan(taxConfig)
	if err != nil {
		return nil, err
	}

	return taxConfig, nil
}
