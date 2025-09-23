#!/bin/bash

echo "Testing JSON authorization error responses..."

# Test 1: Invalid signin credentials (should return JSON)
echo "Test 1: Testing invalid signin credentials..."
response=$(curl -s -X POST http://localhost:8081/api/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrong"}' \
  -w "HTTP_STATUS:%{http_code}")

echo "Response: $response"

# Extract HTTP status
status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$status" = "401" ]; then
  echo "✅ Status code is correct (401)"
else
  echo "❌ Expected status 401, got $status"
fi

# Check if response is valid JSON
echo "$body" | jq . > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "✅ Response is valid JSON"
  echo "JSON response: $(echo "$body" | jq .)"
else
  echo "❌ Response is not valid JSON: $body"
fi

echo ""
echo "Test 2: Testing missing authentication token..."
response2=$(curl -s -X GET http://localhost:8081/api/auth/profile \
  -w "HTTP_STATUS:%{http_code}")

echo "Response: $response2"

# Extract HTTP status
status2=$(echo "$response2" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
body2=$(echo "$response2" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$status2" = "401" ]; then
  echo "✅ Status code is correct (401)"
else
  echo "❌ Expected status 401, got $status2"
fi

# Check if response is valid JSON
echo "$body2" | jq . > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "✅ Response is valid JSON"
  echo "JSON response: $(echo "$body2" | jq .)"
else
  echo "❌ Response is not valid JSON: $body2"
fi