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
# Analytics & Insights APIs
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# These return cached data computed after each receipt upload.
# Upload at least one receipt first before calling these.


# ── 10. Get Analytics (returns cached analytics) ──

curl -s -X GET http://localhost:8080/useranalytics \
  -H "Authorization: Bearer TOKEN" | jq


# ── 11. Get Insights (returns cached LLM insights) ──

curl -s -X GET http://localhost:8080/userinsights \
  -H "Authorization: Bearer TOKEN" | jq


# ── 12. Get Analytics without Auth (expect 400) ──

curl -s -X GET http://localhost:8080/useranalytics | jq


# ── 13. Get Insights without Auth (expect 400) ──

curl -s -X GET http://localhost:8080/userinsights | jq


# ── 14. Get Analytics with Invalid Token (expect 400) ──

curl -s -X GET http://localhost:8080/useranalytics \
  -H "Authorization: Bearer invalidtoken123" | jq


# ── 15. Get Insights with Invalid Token (expect 400) ──

curl -s -X GET http://localhost:8080/userinsights \
  -H "Authorization: Bearer invalidtoken123" | jq
