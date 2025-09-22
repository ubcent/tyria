#!/bin/bash

# Edge.link Proxy Demo Script
# This script demonstrates the key features of the proxy service

BASE_URL="http://localhost:8080"
API_KEY="demo-key-12345"

echo "🚀 Edge.link Proxy Demo"
echo "======================="
echo ""

# Check if server is running
echo "1. Health Check"
echo "---------------"
curl -s "${BASE_URL}/api/health" | jq .
echo ""

# Test public endpoint (no auth required)
echo "2. Public API Proxy (JSONPlaceholder)"
echo "------------------------------------"
echo "GET /api/v1/posts/1"
curl -s "${BASE_URL}/api/v1/posts/1" | jq .
echo ""

# Test the same request again to show caching
echo "3. Cache Hit Test (same request)"
echo "-------------------------------"
echo "GET /api/v1/posts/1 (should be from cache)"
time curl -s "${BASE_URL}/api/v1/posts/1" | jq .title
echo ""

# Test authenticated endpoint
echo "4. Authenticated API (httpbin.org)"
echo "----------------------------------"
echo "GET /api/secure/ip (with API key)"
curl -s -H "Authorization: Bearer ${API_KEY}" "${BASE_URL}/api/secure/ip" | jq .
echo ""

# Test authentication failure
echo "5. Authentication Test (no API key)"
echo "----------------------------------"
echo "GET /api/secure/ip (without API key - should fail)"
curl -s "${BASE_URL}/api/secure/ip"
echo ""

# Show current statistics
echo "6. Proxy Statistics"
echo "------------------"
curl -s "${BASE_URL}/api/stats" | jq .
echo ""

# Show cache statistics
echo "7. Cache Statistics"
echo "------------------"
curl -s "${BASE_URL}/api/cache/stats" | jq .
echo ""

# Show API key information
echo "8. API Key Management"
echo "--------------------"
curl -s "${BASE_URL}/api/auth/keys" | jq .
echo ""

# Show rate limiting stats
echo "9. Rate Limiting Statistics"
echo "--------------------------"
curl -s "${BASE_URL}/api/ratelimit/stats" | jq .
echo ""

# Show metrics from metrics server
echo "10. Metrics Endpoint"
echo "-------------------"
curl -s "http://localhost:9090/metrics" | jq .total_requests,.proxied_requests,.cached_requests
echo ""

echo "✅ Demo complete!"
echo ""
echo "Try these additional tests:"
echo "- Rate limiting: Send many requests quickly"
echo "- Different API keys: Use 'admin-key-67890'"
echo "- Cache clearing: POST to /api/cache/clear"
echo "- Custom routes: Add your own in config.yaml"