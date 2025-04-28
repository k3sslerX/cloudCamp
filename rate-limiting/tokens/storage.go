package tokens

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(pool *pgxpool.Pool) *DBStorage {
	return &DBStorage{pool: pool}
}

func (s *DBStorage) LoadConfigs(rl *RateLimiter) error {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `
		SELECT api_key, capacity, rate, frequency 
		FROM api_rate_limits
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			apiKey    string
			capacity  int
			rate      int
			frequency time.Duration
		)
		if err := rows.Scan(&apiKey, &capacity, &rate, &frequency); err != nil {
			return err
		}
		rl.SetClientConfig(apiKey, ClientConfig{
			Capacity:  capacity,
			Rate:      rate,
			Frequency: frequency,
		})
	}

	return rows.Err()
}

func (s *DBStorage) SaveConfig(ctx context.Context, apiKey string, config ClientConfig) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_rate_limits (api_key, capacity, rate, frequency) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (api_key) DO UPDATE 
		SET capacity = EXCLUDED.capacity,
			rate = EXCLUDED.rate,
			frequency = EXCLUDED.frequency
	`, apiKey, config.Capacity, config.Rate, config.Frequency)
	return err
}

func (s *DBStorage) ValidateAPIKey(ctx context.Context, apiKey string) (bool, error) {
	var valid bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM api_keys WHERE key = $1 AND active = TRUE)
	`, apiKey).Scan(&valid)
	return valid, err
}
