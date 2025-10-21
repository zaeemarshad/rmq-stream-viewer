#!/bin/bash
# Verification script for RabbitMQ Stream Viewer

set -e

echo "=== RabbitMQ Stream Viewer - Verification Script ==="
echo ""

# Check Go version
echo "✓ Checking Go version..."
go version

# Check Node version
echo "✓ Checking Node version..."
node --version

# Run Go tests
echo ""
echo "✓ Running Go tests..."
go test ./...

# Check test coverage
echo ""
echo "✓ Checking test coverage..."
go test -cover ./... | grep -E "coverage:|ok"

# Build backend
echo ""
echo "✓ Building backend..."
go build -o bin/server cmd/server/main.go
go build -o bin/publisher cmd/publisher/main.go
ls -lh bin/

# Build frontend
echo ""
echo "✓ Building frontend..."
cd web
npm run build
ls -lh dist/

echo ""
echo "=== Verification Complete ==="
echo ""
echo "✅ All checks passed!"
echo ""
echo "Next steps:"
echo "  1. Update config.yaml with your RabbitMQ connection details"
echo "  2. Run './bin/server -config config.yaml' to start the server"
echo "  3. Run './bin/publisher' to generate test messages"
echo "  4. Open http://localhost:8080 in your browser"
echo ""

