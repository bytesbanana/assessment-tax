package postgres

import (
	"time"
)

type TaxConfig struct {
	ID        int        `postgres:"id"`
	Key       string     `postgres:"key"`
	Name      string     `postgres:"name"`
	Value     float64    `postgres:"value"`
	CreatedAt *time.Time `postgres:"created_at"`
	CreatedBy *string    `postgres:"created_by"`
	UpdatedAt *time.Time `postgres:"updated_at"`
	UpdatedBy *string    `postgres:"updated_by"`
}

func (p *Postgres) GetTaxConfig(key string) (*TaxConfig, error) {
	var taxConfig *TaxConfig

	row := p.Db.QueryRow("SELECT * FROM tax_configs WHERE key = $1", key)

	err := row.Scan(taxConfig)
	if err != nil {
		return nil, err
	}

	return taxConfig, nil
}

func (p *Postgres) SetTaxConfig(key string, value float64) (*TaxConfig, error) {
	row := p.Db.QueryRow("UPDATE tax_configs SET value = $1 WHERE key = $2 returning *", value, key)
	var config TaxConfig
	var sCreateAt *string
	var sUpdatedAt *string
	err := row.Scan(&config.ID, &config.Key, &config.Name, &config.Value, &config.CreatedBy, &sCreateAt, &config.UpdatedBy, &sUpdatedAt)
	if err != nil {
		return nil, err
	}

	createAt, err := p.dateStrToTime(sCreateAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := p.dateStrToTime(sUpdatedAt)
	if err != nil {
		return nil, err
	}

	config.CreatedAt = createAt
	config.UpdatedAt = updatedAt

	return &config, nil
}

func (p *Postgres) dateStrToTime(dateStr *string) (*time.Time, error) {

	if dateStr != nil {
		date, err := time.Parse(time.RFC3339, *dateStr)
		if err != nil {
			return nil, err
		}

		return &date, nil
	}
	return nil, nil
}
