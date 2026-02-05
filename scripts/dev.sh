#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
INFRA_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.integration.yml"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BACKEND_SERVICES=(
    "auth-service"
    "client-query-service"
    "inventory-service"
    "order-service"
    "product-service"
    "api-gateway"
)

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    cat <<EOF
Usage: $0 <start|stop|restart|status|check|logs> [target]

Targets:
  environment|all       Full local environment (infra + backend + frontend)
  infra                 Infrastructure only (MongoDB + Redis + NATS via Docker)
  backend               All backend services
  frontend              Frontend dev server
  auth-service
  client-query-service
  inventory-service
  order-service
  product-service
  api-gateway

Examples:
  $0 start
  $0 start environment
  $0 start infra
  $0 start api-gateway
  $0 restart backend
  $0 logs frontend
EOF
}

is_backend_service() {
    case "$1" in
        auth-service|client-query-service|inventory-service|order-service|product-service|api-gateway) return 0 ;;
        *) return 1 ;;
    esac
}

service_dir() {
    case "$1" in
        auth-service) echo "$PROJECT_ROOT/cmd/auth-service" ;;
        client-query-service) echo "$PROJECT_ROOT/cmd/client-query-service" ;;
        inventory-service) echo "$PROJECT_ROOT/cmd/inventory-service" ;;
        order-service) echo "$PROJECT_ROOT/cmd/order-service" ;;
        product-service) echo "$PROJECT_ROOT/cmd/product-service" ;;
        api-gateway) echo "$PROJECT_ROOT/cmd/api-gateway" ;;
        *) return 1 ;;
    esac
}

service_port() {
    case "$1" in
        auth-service) echo "8081" ;;
        client-query-service) echo "8082" ;;
        inventory-service) echo "8084" ;;
        order-service) echo "8086" ;;
        product-service) echo "8085" ;;
        api-gateway) echo "8080" ;;
        frontend) echo "5173" ;;
        *) return 1 ;;
    esac
}

pid_file() {
    echo "/tmp/erp-$1.pid"
}

log_file() {
    echo "/tmp/erp-$1.log"
}

service_health_url() {
    local port
    port="$(service_port "$1")"
    echo "http://localhost:${port}/health"
}

is_port_listening() {
    local port="$1"
    if command -v lsof >/dev/null 2>&1; then
        lsof -iTCP:"$port" -sTCP:LISTEN -n -P >/dev/null 2>&1
    elif command -v nc >/dev/null 2>&1; then
        nc -z localhost "$port" >/dev/null 2>&1
    else
        return 1
    fi
}

wait_for_url() {
    local url="$1"
    local timeout="${2:-30}"
    local i=0
    while [ "$i" -lt "$timeout" ]; do
        if curl -fsS "$url" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        i=$((i + 1))
    done
    return 1
}

is_managed_running() {
    local service="$1"
    local file
    file="$(pid_file "$service")"
    if [ ! -f "$file" ]; then
        return 1
    fi

    local pid
    pid="$(cat "$file" 2>/dev/null || true)"
    if [ -z "$pid" ]; then
        return 1
    fi

    kill -0 "$pid" >/dev/null 2>&1
}

run_infra_compose() {
    if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
        docker compose -f "$INFRA_COMPOSE_FILE" "$@"
        return 0
    fi
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose -f "$INFRA_COMPOSE_FILE" "$@"
        return 0
    fi

    log_error "docker compose/docker-compose is not available"
    return 1
}

start_infra() {
    log_info "Starting infrastructure from $INFRA_COMPOSE_FILE..."
    run_infra_compose up -d
}

stop_infra() {
    log_info "Stopping infrastructure..."
    run_infra_compose down
}

status_infra() {
    log_info "Infrastructure status:"
    if ! run_infra_compose ps; then
        log_warn "Unable to read compose status, falling back to local port checks"
        if is_port_listening 27017; then log_info "MongoDB listening on 27017"; else log_warn "MongoDB not listening on 27017"; fi
        if is_port_listening 6379; then log_info "Redis listening on 6379"; else log_warn "Redis not listening on 6379"; fi
        if is_port_listening 4222; then log_info "NATS listening on 4222"; else log_warn "NATS not listening on 4222"; fi
    fi
}

start_backend_service() {
    local service="$1"
    if ! is_backend_service "$service"; then
        log_error "Unknown backend service: $service"
        return 1
    fi

    if is_managed_running "$service"; then
        log_warn "$service is already running (managed by this script)"
        return 0
    fi

    local port
    port="$(service_port "$service")"
    if is_port_listening "$port"; then
        log_warn "$service port $port is already in use"
        return 0
    fi

    local dir
    dir="$(service_dir "$service")"
    local file
    file="$(log_file "$service")"
    local pidpath
    pidpath="$(pid_file "$service")"

    log_info "Starting $service on port $port..."
    (
        cd "$dir"
        nohup go run main.go >"$file" 2>&1 &
        echo $! >"$pidpath"
    )

    if wait_for_url "$(service_health_url "$service")" 30; then
        log_info "$service started successfully"
    else
        log_error "$service did not become healthy in time"
        log_error "Check logs: tail -f $file"
        return 1
    fi
}

stop_backend_service() {
    local service="$1"
    local pidpath
    pidpath="$(pid_file "$service")"

    if [ ! -f "$pidpath" ]; then
        log_warn "$service is not managed by this script (missing $pidpath)"
        return 0
    fi

    local pid
    pid="$(cat "$pidpath" 2>/dev/null || true)"
    if [ -z "$pid" ] || ! kill -0 "$pid" >/dev/null 2>&1; then
        log_warn "$service process is not running"
        rm -f "$pidpath"
        return 0
    fi

    log_info "Stopping $service (PID $pid)..."
    kill "$pid" >/dev/null 2>&1 || true

    local i=0
    while [ "$i" -lt 10 ]; do
        if ! kill -0 "$pid" >/dev/null 2>&1; then
            break
        fi
        sleep 1
        i=$((i + 1))
    done

    if kill -0 "$pid" >/dev/null 2>&1; then
        log_warn "$service did not stop gracefully, force killing PID $pid"
        kill -9 "$pid" >/dev/null 2>&1 || true
    fi

    rm -f "$pidpath"
    log_info "$service stopped"
}

start_backend() {
    local service
    for service in "${BACKEND_SERVICES[@]}"; do
        start_backend_service "$service"
    done
}

stop_backend() {
    local idx
    for ((idx=${#BACKEND_SERVICES[@]}-1; idx>=0; idx--)); do
        stop_backend_service "${BACKEND_SERVICES[$idx]}"
    done
}

status_backend() {
    log_info "Backend services:"
    local service
    for service in "${BACKEND_SERVICES[@]}"; do
        local port
        port="$(service_port "$service")"
        if is_managed_running "$service"; then
            if wait_for_url "$(service_health_url "$service")" 1; then
                echo -e "${GREEN}[RUNNING]${NC} $service (port $port)"
            else
                echo -e "${YELLOW}[STARTED]${NC} $service (port $port, health check failing)"
            fi
        else
            echo -e "${RED}[STOPPED]${NC} $service (port $port)"
        fi
    done
}

start_frontend() {
    local service="frontend"
    local port
    port="$(service_port "$service")"
    local file
    file="$(log_file "$service")"
    local pidpath
    pidpath="$(pid_file "$service")"

    if is_managed_running "$service"; then
        log_warn "frontend is already running (managed by this script)"
        return 0
    fi

    if is_port_listening "$port"; then
        log_warn "frontend port $port is already in use"
        return 0
    fi

    log_info "Starting frontend on port $port..."
    (
        cd "$FRONTEND_DIR"
        nohup npm run dev -- --host 0.0.0.0 --port "$port" >"$file" 2>&1 &
        echo $! >"$pidpath"
    )

    if wait_for_url "http://localhost:${port}" 30; then
        log_info "frontend started successfully"
    else
        log_error "frontend did not become reachable in time"
        log_error "Check logs: tail -f $file"
        return 1
    fi
}

stop_frontend() {
    stop_backend_service "frontend"
}

status_frontend() {
    local port
    port="$(service_port frontend)"
    if is_managed_running "frontend"; then
        if wait_for_url "http://localhost:${port}" 1; then
            echo -e "${GREEN}[RUNNING]${NC} frontend (port $port)"
        else
            echo -e "${YELLOW}[STARTED]${NC} frontend (port $port, endpoint unreachable)"
        fi
    else
        echo -e "${RED}[STOPPED]${NC} frontend (port $port)"
    fi
}

start_target() {
    local target="$1"
    case "$target" in
        environment|all)
            start_infra
            start_backend
            start_frontend
            ;;
        infra) start_infra ;;
        backend) start_backend ;;
        frontend) start_frontend ;;
        auth-service|client-query-service|inventory-service|order-service|product-service|api-gateway)
            start_backend_service "$target"
            ;;
        *)
            log_error "Unknown start target: $target"
            usage
            return 1
            ;;
    esac
}

stop_target() {
    local target="$1"
    case "$target" in
        environment|all)
            stop_frontend
            stop_backend
            stop_infra
            ;;
        infra) stop_infra ;;
        backend) stop_backend ;;
        frontend) stop_frontend ;;
        auth-service|client-query-service|inventory-service|order-service|product-service|api-gateway)
            stop_backend_service "$target"
            ;;
        *)
            log_error "Unknown stop target: $target"
            usage
            return 1
            ;;
    esac
}

status_target() {
    local target="$1"
    case "$target" in
        environment|all)
            status_infra
            status_backend
            status_frontend
            ;;
        infra) status_infra ;;
        backend) status_backend ;;
        frontend) status_frontend ;;
        auth-service|client-query-service|inventory-service|order-service|product-service|api-gateway)
            status_backend | grep "$target" || true
            ;;
        *)
            log_error "Unknown status target: $target"
            usage
            return 1
            ;;
    esac
}

show_logs() {
    local target="${1:-api-gateway}"
    case "$target" in
        infra)
            run_infra_compose logs -f
            ;;
        frontend|auth-service|client-query-service|inventory-service|order-service|product-service|api-gateway)
            local file
            file="$(log_file "$target")"
            if [ ! -f "$file" ]; then
                log_error "No log file found for $target at $file"
                return 1
            fi
            tail -f "$file"
            ;;
        *)
            log_error "Unknown logs target: $target"
            usage
            return 1
            ;;
    esac
}

check_infrastructure() {
    status_infra
}

ACTION="${1:-start}"
TARGET="${2:-environment}"

case "$ACTION" in
    start) start_target "$TARGET" ;;
    stop) stop_target "$TARGET" ;;
    restart)
        stop_target "$TARGET"
        sleep 1
        start_target "$TARGET"
        ;;
    status) status_target "$TARGET" ;;
    check) check_infrastructure ;;
    logs) show_logs "${2:-api-gateway}" ;;
    help|-h|--help) usage ;;
    *)
        log_error "Unknown command: $ACTION"
        usage
        exit 1
        ;;
esac
