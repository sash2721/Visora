package repositories

import (
	"Backend/models"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadRepository struct {
	DB *pgxpool.Pool
}

func NewUploadRepository(db *pgxpool.Pool) *UploadRepository {
	return &UploadRepository{DB: db}
}

// StoreReceipt inserts the receipt and its items in a single transaction.
func (repo *UploadRepository) StoreReceipt(userID string, receipt models.GenAIUploadResponse, imageURL string) (string, error) {
	tx, err := repo.DB.Begin(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background()) // no-op if already committed

	// insert receipt and get back the generated UUID
	receiptQuery := `
		INSERT INTO receipts (user_id, merchant, date, total_amount, currency, confidence_score, image_url, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'scan')
		RETURNING id`

	var receiptID string
	err = tx.QueryRow(context.Background(), receiptQuery,
		userID,
		receipt.Merchant,
		receipt.Date,
		receipt.TotalAmount,
		receipt.Currency,
		receipt.ConfidenceScore,
		imageURL,
	).Scan(&receiptID)

	if err != nil {
		return "", fmt.Errorf("failed to insert receipt: %w", err)
	}

	// insert each item, looking up category_id by name
	itemQuery := `
		INSERT INTO items (receipt_id, name, price, category_id, quantity)
		VALUES ($1, $2, $3, (SELECT id FROM categories WHERE name = $4), $5)`

	for _, item := range receipt.Items {
		_, err = tx.Exec(context.Background(), itemQuery,
			receiptID,
			item.Name,
			item.Price,
			item.Category,
			item.Quantity,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert item %s: %w", item.Name, err)
		}
	}

	// commit the transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Receipt stored in DB",
		slog.String("ReceiptID", receiptID),
		slog.String("UserID", userID),
		slog.Int("ItemCount", len(receipt.Items)),
	)

	return receiptID, nil
}

// StoreManualExpense inserts a manually entered expense and its items in a single transaction.
func (repo *UploadRepository) StoreManualExpense(userID string, expense models.ManualExpenseRequest, totalAmount float64) (string, error) {
	tx, err := repo.DB.Begin(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	receiptQuery := `
		INSERT INTO receipts (user_id, merchant, date, total_amount, currency, confidence_score, image_url, source)
		VALUES ($1, $2, $3, $4, $5, 1.0, '', 'manual')
		RETURNING id`

	var receiptID string
	err = tx.QueryRow(context.Background(), receiptQuery,
		userID,
		expense.Merchant,
		expense.Date,
		totalAmount,
		expense.Currency,
	).Scan(&receiptID)

	if err != nil {
		return "", fmt.Errorf("failed to insert manual expense: %w", err)
	}

	itemQuery := `
		INSERT INTO items (receipt_id, name, price, category_id, quantity)
		VALUES ($1, $2, $3, (SELECT id FROM categories WHERE name = $4), $5)`

	for _, item := range expense.Items {
		_, err = tx.Exec(context.Background(), itemQuery,
			receiptID,
			item.Name,
			item.Price,
			item.Category,
			item.Quantity,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert item %s: %w", item.Name, err)
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Manual expense stored in DB",
		slog.String("ReceiptID", receiptID),
		slog.String("UserID", userID),
		slog.Int("ItemCount", len(expense.Items)),
	)

	return receiptID, nil
}
