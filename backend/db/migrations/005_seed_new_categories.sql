-- Add new categories for small/receiptless purchases and discounts
INSERT INTO categories (name) VALUES
    ('Snacks & Beverages'),
    ('Parking'),
    ('Tips & Service Charges'),
    ('Discounts & Cashback')
ON CONFLICT (name) DO NOTHING;
