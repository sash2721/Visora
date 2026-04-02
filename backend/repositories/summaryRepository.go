package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReceiptRow represents a single flat row from the receipts+items+categories join.
type ReceiptRow struct {
	ReceiptID    string
	Merchant     string
	Date         time.Time
	TotalAmount  float64
	Currency     string
	ItemName     string
	ItemPrice    float64
	ItemQuantity int
	CategoryName string
}

type SummaryRepository struct {
	DB *pgxpool.Pool
}

func NewSummaryRepository(db *pgxpool.Pool) *SummaryRepository {
	return &SummaryRepository{DB: db}
}

// GetUserReceipts fetches all receipts with their items and category names for a given user.
func (repo *SummaryRepository) GetUserReceipts(userID string) ([]ReceiptRow, error) {
	query := `
		SELECT
			r.id          AS receipt_id,
			r.merchant,
			r.date,
			r.total_amount,
			r.currency,
			i.name        AS item_name,
			i.price       AS item_price,
			i.quantity    AS item_quantity,
			c.name        AS category_name
		FROM receipts r
		JOIN items i      ON i.receipt_id = r.id
		JOIN categories c ON c.id = i.category_id
		WHERE r.user_id = $1
		ORDER BY r.date DESC, r.id, i.name`

	rows, err := repo.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query receipts: %w", err)
	}
	defer rows.Close()

	var results []ReceiptRow
	for rows.Next() {
		var row ReceiptRow
		err := rows.Scan(
			&row.ReceiptID, &row.Merchant, &row.Date, &row.TotalAmount, &row.Currency,
			&row.ItemName, &row.ItemPrice, &row.ItemQuantity, &row.CategoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan receipt row: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating receipt rows: %w", err)
	}

	slog.Info("Fetched user receipt rows", slog.String("UserID", userID), slog.Int("RowCount", len(results)))
	return results, nil
}

// UpsertAnalytics inserts or updates cached analytics for a user+period.
func (repo *SummaryRepository) UpsertAnalytics(userID string, totalSpent float64, currency string, categoryBreakdown []byte, dailySpending []byte, period string) error {
	query := `
		INSERT INTO user_analytics (user_id, total_spent, currency, category_breakdown, daily_spending, period, computed_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, period)
		DO UPDATE SET
			total_spent = EXCLUDED.total_spent,
			currency = EXCLUDED.currency,
			category_breakdown = EXCLUDED.category_breakdown,
			daily_spending = EXCLUDED.daily_spending,
			computed_at = NOW()`

	_, err := repo.DB.Exec(context.Background(), query, userID, totalSpent, currency, categoryBreakdown, dailySpending, period)
	if err != nil {
		return fmt.Errorf("failed to upsert analytics: %w", err)
	}
	return nil
}

// UpsertInsights inserts or updates cached insights for a user+period.
func (repo *SummaryRepository) UpsertInsights(userID string, summary string, warnings []byte, period string) error {
	query := `
		INSERT INTO user_insights (user_id, summary, warnings, period, computed_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, period)
		DO UPDATE SET
			summary = EXCLUDED.summary,
			warnings = EXCLUDED.warnings,
			computed_at = NOW()`

	_, err := repo.DB.Exec(context.Background(), query, userID, summary, warnings, period)
	if err != nil {
		return fmt.Errorf("failed to upsert insights: %w", err)
	}
	return nil
}

// CachedAnalytics represents a row from the user_analytics table.
type CachedAnalytics struct {
	TotalSpent        float64
	Currency          string
	CategoryBreakdown []byte // raw JSONB
	DailySpending     []byte // raw JSONB
	Period            string
	ComputedAt        time.Time
}

// CachedInsights represents a row from the user_insights table.
type CachedInsights struct {
	Summary    string
	Warnings   []byte // raw JSONB
	Period     string
	ComputedAt time.Time
}

// GetCachedAnalytics fetches the cached analytics for a user+period. Returns nil if not found.
func (repo *SummaryRepository) GetCachedAnalytics(userID string, period string) (*CachedAnalytics, error) {
	query := `SELECT total_spent, currency, category_breakdown, daily_spending, period, computed_at
		FROM user_analytics WHERE user_id = $1 AND period = $2`

	var result CachedAnalytics
	err := repo.DB.QueryRow(context.Background(), query, userID, period).Scan(
		&result.TotalSpent, &result.Currency, &result.CategoryBreakdown,
		&result.DailySpending, &result.Period, &result.ComputedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch cached analytics: %w", err)
	}

	return &result, nil
}

// GetCachedInsights fetches the cached insights for a user+period. Returns nil if not found.
func (repo *SummaryRepository) GetCachedInsights(userID string, period string) (*CachedInsights, error) {
	query := `SELECT summary, warnings, period, computed_at
		FROM user_insights WHERE user_id = $1 AND period = $2`

	var result CachedInsights
	err := repo.DB.QueryRow(context.Background(), query, userID, period).Scan(
		&result.Summary, &result.Warnings, &result.Period, &result.ComputedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch cached insights: %w", err)
	}

	return &result, nil
}
