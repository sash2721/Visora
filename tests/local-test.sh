#!/bin/bash

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Local Test Commands - Copy & Paste to run
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Replace TOKEN with the actual JWT token from signup/login response
# Replace /path/to/receipt.jpg with your actual image path


# ── 1. Health Check ──

curl -s http://localhost:8080/health | jq


# ── 2. Signup ──

curl -s -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "Test@1234",
    "country": "IN"
  }' | jq


# ── 3. Login ──

curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "Test@1234"
  }' | jq


# ── 4. Upload Receipt (requires GenAI service running) ──

curl -s -X POST http://localhost:8080/uploadreceipt \
  -H "Authorization: Bearer TOKEN" \
  -F "image=@/path/to/receipt.jpg" \
  -F "currency=INR" | jq


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Error Cases
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


# ── 5. Duplicate Signup (expect 400) ──

curl -s -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "Test@1234",
    "country": "IN"
  }' | jq


# ── 6. Login with Wrong Password (expect 400) ──

curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "WrongPass123"
  }' | jq


# ── 7. Upload without Auth Token (expect 400) ──

curl -s -X POST http://localhost:8080/uploadreceipt \
  -F "image=@/path/to/receipt.jpg" \
  -F "currency=INR" | jq


# ── 8. Upload with Invalid Token (expect 400) ──

curl -s -X POST http://localhost:8080/uploadreceipt \
  -H "Authorization: Bearer invalidtoken123" \
  -F "image=@/path/to/receipt.jpg" \
  -F "currency=INR" | jq


# ── 9. Upload Missing Currency (expect 400) ──

curl -s -X POST http://localhost:8080/uploadreceipt \
  -H "Authorization: Bearer TOKEN" \
  -F "image=@/path/to/receipt.jpg" | jq


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Manual Expense Entry
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


# ── 10. Add Manual Expense ──

curl -s -X POST http://localhost:8080/manualexpense \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant": "Street Food Stall",
    "date": "2026-04-02",
    "currency": "INR",
    "items": [
      { "name": "Samosa", "price": 20, "quantity": 2, "category": "Snacks & Beverages" },
      { "name": "Chai", "price": 15, "quantity": 1, "category": "Snacks & Beverages" }
    ]
  }' | jq


# ── 11. Add Manual Expense - Parking ──

curl -s -X POST http://localhost:8080/manualexpense \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant": "Mall Parking",
    "date": "2026-04-02",
    "currency": "INR",
    "items": [
      { "name": "Parking - 2 hours", "price": 50, "quantity": 1, "category": "Parking" }
    ]
  }' | jq


# ── 12. Manual Expense without Auth (expect 400) ──

curl -s -X POST http://localhost:8080/manualexpense \
  -H "Content-Type: application/json" \
  -d '{
    "merchant": "Test",
    "date": "2026-04-02",
    "currency": "INR",
    "items": [{ "name": "Item", "price": 10, "quantity": 1, "category": "Miscellaneous" }]
  }' | jq


# ── 13. Manual Expense Missing Fields (expect 400) ──

curl -s -X POST http://localhost:8080/manualexpense \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant": "Test"
  }' | jq


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Analytics & Insights APIs
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# These return cached data computed after each receipt upload.
# Upload at least one receipt first before calling these.


# ── 14. Get Analytics (returns cached analytics) ──

curl -s -X GET http://localhost:8080/useranalytics \
  -H "Authorization: Bearer TOKEN" | jq


# ── 15. Get Insights (returns cached LLM insights) ──

curl -s -X GET http://localhost:8080/userinsights \
  -H "Authorization: Bearer TOKEN" | jq


# ── 16. Get Analytics without Auth (expect 400) ──

curl -s -X GET http://localhost:8080/useranalytics | jq


# ── 17. Get Insights without Auth (expect 400) ──

curl -s -X GET http://localhost:8080/userinsights | jq


# ── 18. Get Analytics with Invalid Token (expect 400) ──

curl -s -X GET http://localhost:8080/useranalytics \
  -H "Authorization: Bearer invalidtoken123" | jq


# ── 19. Get Insights with Invalid Token (expect 400) ──

curl -s -X GET http://localhost:8080/userinsights \
  -H "Authorization: Bearer invalidtoken123" | jq
