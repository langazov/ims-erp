#!/bin/bash

# ERP System Development Environment
# Starts all services with proper configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if infrastructure is running
check_infrastructure() {
    log_info "Checking infrastructure..."

    # Check MongoDB
    if ! curl -s mongo://localhost:27017 > /dev/null 2>&1; then
        log_warn "MongoDB not running. Start with: docker-compose -f $PROJECT_ROOT/docker-compose.integration.yml up -d"
    else
        log_info "MongoDB is running"
    fi

    # Check Redis
    if ! redis-cli ping > /dev/null 2>&1; then
        log_warn "Redis not running"
    else
        log_info "Redis is running"
    fi

    # Check NATS
    if ! curl -s nats://localhost:4222 > /dev/null 2>&1; then
        log_warn "NATS not running"
    else
        log_info "NATS is running"
    fi
}

# Start a service in the background
start_service() {
    local service_name=$1
    local service_dir=$2
    local port=$3
    local log_file="/tmp/erp-$service_name.log"

    # Check if already running
    if lsof -i :$port > /dev/null 2>&1; then
        log_warn "$service_name is already running on port $port"
        return 0
    fi

    log_info "Starting $service_name on port $port..."
    cd "$service_dir"
    go run main.go > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "/tmp/erp-$service_name.pid"

    # Wait for service to start
    sleep 3

    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        log_info "$service_name started successfully (PID: $pid)"
    else
        log_error "Failed to start $service_name"
        log_error "Check logs: tail -f $log_file"
    fi
}

# Stop all services
stop_all() {
    log_info "Stopping all services..."

    for pid_file in /tmp/erp-*.pid; do
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if kill -0 $pid 2>/dev/null; then
                kill $pid 2>/dev/null || true
                log_info "Stopped process $pid"
            fi
            rm -f "$pid_file"
        fi
    done

    log_info "All services stopped"
}

# Start all services
start_all() {
    log_info "Starting ERP System Development Environment..."

    # Check infrastructure first
    check_infrastructure

    log_info "Starting backend services..."

    # Start services in order (dependencies first)
    start_service "auth-service" "$PROJECT_ROOT/cmd/auth-service" 8081
    start_service "client-query-service" "$PROJECT_ROOT/cmd/client-query-service" 8082
    start_service "inventory-service" "$PROJECT_ROOT/cmd/inventory-service" 8084
    start_service "order-service" "$PROJECT_ROOT/cmd/order-service" 8086
    start_service "product-service" "$PROJECT_ROOT/cmd/product-service" 8085
    start_service "api-gateway" "$PROJECT_ROOT/cmd/api-gateway" 8080

    log_info ""
    log_info "========================================="
    log_info "All services started!"
    log_info "========================================="
    log_info ""
    log_info "API Gateway: http://localhost:8080"
    log_info "Health:      http://localhost:8080/health"
    log_info "Ready:       http://localhost:8080/ready"
    log_info ""
    log_info "Frontend:    http://localhost:5173 (run: cd frontend && npm run dev)"
    log_info ""
    log_info "Service ports:"
    log_info "  - API Gateway:        8080"
    log_info "  - Auth Service:       8081"
    log_info "  - Client Query:       8082"
    log_info "  - Product Service:    8085"
    log_info "  - Inventory Service:  8084"
    log_info "  - Order Service:      8086"
    log_info ""
}

# Status of all services
status() {
    echo "ERP System Services Status"
    echo "=========================="
    echo ""

    local services=(
        "api-gateway:8080"
        "auth-service:8081"
        "client-query-service:8082"
        "inventory-service:8084"
        "product-service:8085"
        "order-service:8086"
    )

    for svc in "${services[@]}"; do
        local name="${svc%%:*}"
        local port="${svc##*:}"

        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            local status=$(curl -s "http://localhost:$port/health" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}[RUNNING]${NC} $name (port $port) - $status"
        else
            echo -e "${RED}[STOPPED]${NC} $name (port $port)"
        fi
    done
}

# Main command handling
case "${1:-start}" in
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    restart)
        stop_all
        sleep 2
        start_all
        ;;
    status)
        status
        ;;
    check)
        check_infrastructure
        status
        ;;
    logs)
        service="${2:-api-gateway}"
        tail -f "/tmp/erp-$service.log"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|check|logs [service]}"
        exit 1
        ;;
esac
