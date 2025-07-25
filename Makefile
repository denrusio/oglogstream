.PHONY: help dev prod down build logs test benchmark clean status

# Default target
help:
	@echo "OgLogStream - Load Balanced Log Processing Platform"
	@echo ""
	@echo "Available commands:"
	@echo "  dev         - Start development environment (single instances)"
	@echo "  prod        - Start production environment (load balanced replicas)"
	@echo "  down        - Stop all containers"
	@echo "  build       - Build all Docker images"
	@echo "  logs        - Show logs from all services"
	@echo "  status      - Show status of all services"
	@echo "  test        - Run all tests"
	@echo "  benchmark   - Run performance benchmarks"
	@echo "  clean       - Clean up containers and images"
	@echo "  stats       - Show HAProxy load balancer stats (prod only)"

# Development environment (single instances)
dev:
	@echo "🚀 Starting development environment..."
	docker compose up --build -d
	@echo ""
	@echo "✅ Development environment started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Ingestion API: http://localhost:8080"
	@echo "Query API: http://localhost:8081"
	@echo "ClickHouse: http://localhost:8123"

# Production environment (load balanced)
prod:
	@echo "🚀 Starting production environment with load balancer..."
	docker compose -f docker-compose.prod.yml up --build -d
	@echo ""
	@echo "✅ Production environment started!"
	@echo "Load Balanced Frontend: http://localhost:80"
	@echo "HAProxy Stats: http://localhost:8404/stats"
	@echo ""
	@echo "Services running:"
	@echo "  - 2x Frontend instances"
	@echo "  - 3x Ingestion API instances"
	@echo "  - 3x Query API instances"
	@echo "  - 3x Processing Service instances"
	@echo "  - 1x HAProxy Load Balancer"

# Stop all containers
down:
	@echo "🛑 Stopping all services..."
	docker compose down
	docker compose -f docker-compose.prod.yml down
	@echo "✅ All services stopped!"

# Build all images
build:
	@echo "🔨 Building all Docker images..."
	docker compose build
	docker compose -f docker-compose.prod.yml build
	@echo "✅ All images built!"

# Show logs
logs:
	docker compose logs -f

# Show container status
status:
	@echo "📊 Service Status:"
	@echo ""
	docker compose ps
	@echo ""
	@echo "📊 Production Service Status:"
	@echo ""
	docker compose -f docker-compose.prod.yml ps

# Run tests
test:
	@echo "🧪 Running backend tests..."
	docker compose exec ingestion-api go test -v ./... || true
	docker compose exec query-api go test -v ./... || true
	docker compose exec processing-svc go test -v ./... || true
	@echo "✅ Tests completed!"

# Run benchmarks
benchmark:
	@echo "⚡ Running performance benchmarks..."
	docker compose exec query-api go test -bench=. -benchtime=5s || true
	@echo "✅ Benchmarks completed!"

# Clean up everything
clean:
	@echo "🧹 Cleaning up containers and images..."
	docker compose down -v --remove-orphans
	docker compose -f docker-compose.prod.yml down -v --remove-orphans
	docker system prune -f
	@echo "✅ Cleanup completed!"

# Show HAProxy stats (production only)
stats:
	@echo "📈 HAProxy Load Balancer Stats:"
	@echo "URL: http://localhost:8404/stats"
	@echo ""
	@echo "🌐 Production URLs:"
	@echo "  • Main Page: http://localhost:80"
	@echo "  • API Logs: http://localhost:80/api/logs"
	@echo "  • Health Check: http://localhost:80/health"
	@echo ""
	@echo "📊 Backend Status:"
	@curl -s http://localhost:8404/stats 2>/dev/null | grep -E "(active_up|UP)" | wc -l | xargs echo "  Active servers:" || echo "  HAProxy not running"
	@echo ""
	@echo "💻 Open HAProxy Stats Dashboard:"
	@open http://localhost:8404/stats 2>/dev/null || echo "  Visit: http://localhost:8404/stats"

# Quick smoke test
smoke-test:
	@echo "🔥 Running smoke test..."
	@echo "Testing ingestion via load balancer..."
	@curl -X POST http://localhost:80/log -H "Content-Type: application/json" -d '{"level":"info","message":"Smoke test via HAProxy","service":"test"}' && echo " ✅"
	@echo "Testing query via load balancer..."
	@curl -s http://localhost:80/api/logs | head -1 | grep -q "timestamp" && echo "✅ API working" || echo "❌ API failed"
	@echo "Testing frontend..."
	@curl -s http://localhost:3000/ | grep -q "<!DOCTYPE html>" && echo "✅ Frontend serving" || echo "❌ Frontend failed"
	@echo "✅ Smoke test completed!"

# Open frontend in browser
open:
	@echo "🌐 Opening OgLogStream in browser..."
	@open http://localhost:3000 2>/dev/null || echo "Visit: http://localhost:3000"
	@echo ""
	@echo "📊 Production Load Balancer: http://localhost:80"
	@echo "📈 HAProxy Stats: http://localhost:8404/stats"
	@echo "🧪 API Test Page: http://localhost:8000/test-api.html"

# Test API connectivity
test-api:
	@echo "🧪 Testing API connectivity..."
	@echo "📡 Direct API test:"
	@curl -s http://localhost:80/api/logs | head -1 | grep -q "timestamp" && echo "  ✅ API responding" || echo "  ❌ API not responding"
	@echo "📬 Ingestion test:"
	@curl -X POST http://localhost:80/log -H "Content-Type: application/json" -d '{"level":"info","message":"API test","service":"test"}' >/dev/null 2>&1 && echo "  ✅ Ingestion working" || echo "  ❌ Ingestion failed"
	@echo "🔗 CORS test (simulating frontend):"
	@curl -H "Origin: http://localhost:3000" -s http://localhost:80/api/logs | head -1 | grep -q "timestamp" && echo "  ✅ CORS working" || echo "  ❌ CORS failed" 