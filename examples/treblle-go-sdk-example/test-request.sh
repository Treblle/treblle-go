#!/bin/bash

# Replace with your actual ngrok URL
NGROK_URL="YOUR-NGROK-URL.ngrok-free.app"

# Test creating a user with Treblle headers
echo "Creating a user with Treblle tracking headers..."
curl -X POST "${NGROK_URL}/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 5678" \
  -H "X-Trace-ID: test-trace-5678" \
  -d '{"name": "Harry Potter", "email": "harry@example.com"}'

echo -e "\n\nGetting all users with Treblle tracking headers..."
curl -X GET "${NGROK_URL}/api/v1/users" \
  -H "X-User-ID: 5678" \
  -H "X-Trace-ID: test-trace-5678"
