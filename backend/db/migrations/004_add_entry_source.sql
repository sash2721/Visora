-- Add source column to receipts to distinguish between OCR-scanned and manually entered expenses
ALTER TABLE receipts ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'scan';

-- Update existing receipts to 'scan' (they were all uploaded via OCR)
UPDATE receipts SET source = 'scan' WHERE source IS NULL;
