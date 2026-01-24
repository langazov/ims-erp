# AGENTS.md - ERP System Development Guide

This file provides guidelines and commands for AI agents working on the ERP System codebase.

## Build Commands

```bash
# Build all services
make build

# Build specific service (e.g., invoice-service)
make build/invoice-service

# Build with verbose output
go build -v ./...

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/service ./cmd/service
```

## Test Commands

```bash
# Run all tests
make test
go test ./...

# Run tests with coverage
make test-coverage

# Run integration tests (requires infrastructure)
make test-integration

# Run tests in short mode (skips integration tests)
go test -short ./...

# Run a single test file
go test -v ./internal/domain/... -run TestInvoice

# Run a specific test
go test -v ./internal/domain/invoice_test.go -run TestInvoiceAddLine

# Run tests with verbose output and race detection
go test -race -v ./internal/...

# Run tests matching pattern
go test ./... -run "TestPayment|TestProduct"

# Run tests with timeout
go test ./... -timeout 30s

# Run tests excluding certain packages
go test $(go list ./... | grep -v '/internal/integration')
```

## Lint Commands

```bash
# Run linter (requires golangci-lint)
make lint
golangci-lint run ./...

# Run go vet
make vet
go vet ./...

# Format code
make fmt
gofmt -s ./...
gofmt -w file.go  # Format specific file

# Check for unused imports
goimports -l ./...
```

## Code Style Guidelines

### Go Version
- **Required:** Go 1.25.6
- Run `go version` to verify installation

### Imports
- Use standard library imports first, then third-party, then internal
- Group imports with blank lines between groups
- Use goimports or go fmt for automatic formatting

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"

    "github.com/ims-erp/system/internal/domain"
    "github.com/ims-erp/system/pkg/logger"
)
```

### Naming Conventions

**Packages:**
- Use lowercase, single-word names
- Avoid underscores or camelCase
- Be descriptive: `repository`, `messaging`, `tracer`

**Variables:**
- Use camelCase: `clientID`, `invoiceTotal`
- Prefer short names for locals: `i`, `j`, `ctx`
- Avoid underscores in names: `userId` not `user_id`

**Constants:**
- Use camelCase for values: `const MaxRetries = 5`
- Use UPPER_SNAKE for enum-like constants:
```go
const (
    ClientStatusActive     ClientStatus = "active"
    ClientStatusInactive   ClientStatus = "inactive"
)
```

**Types:**
- Use PascalCase: `Client`, `InvoiceStatus`, `PaymentRequest`
- Suffix aggregates with name: `ClientAggregate`, `InvoiceCommand`

**Interfaces:**
- Use -er suffix for single-method interfaces: `Reader`, `Writer`
- Name interfaces after behavior: `PaymentProcessor`, `EventHandler`

**Error Variables:**
- Prefix with package name: `ErrPaymentNotFound`, `ErrInvalidInput`
- Use `var` for error variables, `errors.New`/`fmt.Errorf` for factories

### Types and Declarations

**Structs:**
```go
type Client struct {
    ID        uuid.UUID
    Name      string
    Status    ClientStatus
    CreatedAt time.Time
}
```

**Receiver Methods:**
- Pointer receiver for methods modifying struct: `func (c *Client) Update()`
- Value receiver for read-only methods: `func (c Client) AvailableCredit()`

**Generics:**
- Use when type-agnostic behavior is needed
- Keep constraints minimal and clear

### Error Handling

```go
// Return errors with context
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}

// Custom error types for domain errors
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Check errors with errors.Is/As
if errors.Is(err, ErrNotFound) { ... }
```

### Context Usage

- Pass context as first parameter: `func Do(ctx context.Context, req Request)`
- Use `context.TODO()` when unsure
- Check context cancellation: `select { case <-ctx.Done(): ... }`
- Don't store context in structs

### Testing

**Test File Naming:**
- `*_test.go` suffix
- Parallel to implementation: `invoice.go` → `invoice_test.go`

**Test Functions:**
```go
func TestClient_Update(t *testing.T) {
    // Setup
    client := NewClient(...)

    // Execute
    err := client.Update("new name")

    // Verify
    require.NoError(t, err)
    assert.Equal(t, "new name", client.Name)
}
```

**Test Helpers:**
- Use testify assertions: `assert.Equal`, `require.NoError`
- Table-driven tests for multiple cases:
```go
tests := []struct {
    name    string
    input   string
    want    string
}{
    {"valid", "test@example.com", "test@example.com"},
    {"empty", "", ""},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Logging

- Use structured logging (zap/slog)
- Include correlation IDs for tracing
```go
logger.Info("client created",
    zap.String("client_id", client.ID.String()),
    zap.String("tenant_id", tenantID.String()),
)
```

### Configuration

- Use Viper for configuration management
- Support YAML files and environment variables
- Provide sensible defaults

### Documentation

- Document exported types, functions, and constants
- Use Go doc comments: `// Client represents a customer entity.`
- Keep godoc up to date

## Project Structure

```
cmd/           - Service entry points (one per microservice)
internal/      - Private application code
  domain/      - Domain models and business logic
  commands/    - CQRS command handlers
  queries/     - CQRS query handlers
  events/      - Event handlers
  repository/  - Data access layer
  messaging/   - NATS pub/sub
  config/      - Configuration
  auth/        - Authentication
  rbac/        - Authorization
pkg/           - Public libraries (reusable packages)
api/           - API specifications (OpenAPI)
deployments/   - Kubernetes, Docker, Helm
```

## Common Patterns

**CQRS Pattern:**
- Commands modify state → emit events
- Queries read from projections
- Commands: `CreateInvoice`, `ProcessPayment`
- Queries: `ListInvoices`, `GetClientDetail`

**Event Sourcing:**
- Store events, not state
- Rebuild state by replaying events
- Events in `internal/events/`

**Repository Pattern:**
- Abstraction over data storage
- MongoDB for persistence
- Redis for caching

## Useful Commands

```bash
# Update dependencies
go mod tidy
go mod download

# Add new dependency
go get github.com/new/package@latest

# Remove unused dependencies
go mod tidy

# Generate mocks (requires mockery)
go generate ./...

# Check for outdated dependencies
go list -m -u all
```

## Dependency Management

### Prefer Updating Libraries Over Downgrading Go/Frameworks

When encountering version incompatibilities, **always prefer updating libraries to newer versions** rather than reducing Go version or frontend framework versions.

**Why Update Libraries?**
- Security patches and bug fixes
- Performance improvements
- New features and better APIs
- Long-term maintainability

**Go Version Commitment:**
- **Required:** Go 1.25.6 (minimum)
- Newer Go versions are backward compatible
- Update dependencies to support newer Go versions
- Never reduce Go version to accommodate old libraries

**Workflow for Version Conflicts:**

```bash
# 1. First, try updating the dependency
go get github.com/library/name@latest

# 2. If conflict, update all dependencies
go get -u ./...

# 3. Run go mod tidy to clean up
go mod tidy

# 4. Verify everything still builds
go build ./...

# 5. Run tests to ensure compatibility
go test ./... -short
```

**When You MUST Update:**
- Security vulnerabilities (CVEs)
- Deprecated APIs
- Performance issues
- Missing features needed for implementation

**When to Consider Version Constraints:**
- Only if library has no newer version compatible with Go 1.25.6
- Document why constraint was necessary
- Add TODO to revisit when possible

**Example:**
```bash
# ❌ AVOID: Downgrading Go version for old library
# GO_VERSION=1.20 go build ./...

# ✅ PREFERRED: Update library to compatible version
go get github.com/old-library@v2.0.0
go mod tidy
go build ./...
```

**Frontend Framework Updates:**
- Keep React/Vue/Angular versions current
- Update npm packages regularly: `npm update`
- Use lockfiles for reproducible builds
- Test thoroughly after major version bumps

**Dependency Update Schedule:**
- Security patches: Immediately
- Minor updates: Monthly
- Major updates: Quarterly (with testing)

---

## Key Dependencies

| Package | Purpose |
|---------|---------|
| google/uuid | UUID generation |
| shopspring/decimal | Financial calculations |
| stretchr/testify | Testing assertions |
| spf13/viper | Configuration |
| go.mongodb.org/mongo-driver | MongoDB access |
| nats-io/nats.go | Message bus |
| go-redis/redis | Redis client |
| uber.org/zap | Structured logging |

---

## Concurrency Guidelines

### Prefer Channels Over Mutexes

In Go, **prefer channels for communication between goroutines**. Use mutexes only when protecting simple counters or small structs where channels would add unnecessary complexity.

**When to use channels:**
- Sharing data between goroutines
- Distributing work among workers
- Coordinating asynchronous operations
- Implementing producer-consumer patterns
- Sending signals/cancellation

**When to use mutexes:**
- Simple counters
- Small state protection where channels add overhead
- Cache warming/pre-population

```go
// ✅ GOOD: Use channels for sharing state
type WorkerPool struct {
    tasks   chan Task
    results chan Result
    wg      sync.WaitGroup
}

func (p *WorkerPool) Submit(task Task) {
    p.tasks <- task
}

// ❌ AVOID: Mutex for complex state
type BadCounter struct {
    mu     sync.Mutex
    count  int
    history []int
}

// ✅ GOOD: Channel-based state
type GoodCounter struct {
    count   chan int
    history chan int
}
```

**Worker Pool Pattern:**
```go
func NewWorkerPool(numWorkers int) *WorkerPool {
    tasks := make(chan Task, 100)
    results := make(chan Result, 100)
    
    wp := &WorkerPool{
        tasks:   tasks,
        results: results,
    }
    
    for i := 0; i < numWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
    
    return wp
}

func (p *WorkerPool) worker(id int) {
    defer p.wg.Done()
    for task := range p.tasks {
        p.results <- p.process(task)
    }
}
```

**Context for Cancellation:**
```go
func (p *WorkerPool) Shutdown(ctx context.Context) {
    close(p.tasks)
    p.wg.Wait()
    
    select {
    case <-ctx.Done():
        // Handle timeout
    case <-p.results:
        // Drain results
    }
}
```

---

## Design Patterns

Use established design patterns to solve recurring problems:

### 1. Repository Pattern
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type MongoUserRepository struct {
    collection *mongo.Collection
}
```

### 2. Factory Pattern
```go
type PaymentProcessorFactory interface {
    Create(provider string, config PaymentConfig) (PaymentProcessor, error)
}

func NewPaymentProcessor(provider string, cfg PaymentConfig) (PaymentProcessor, error) {
    switch provider {
    case "stripe":
        return NewStripeProcessor(cfg.APIKey), nil
    case "paypal":
        return NewPayPalProcessor(cfg.ClientID, cfg.ClientSecret), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", provider)
    }
}
```

### 3. Strategy Pattern
```go
type PricingStrategy interface {
    Calculate(items []Item) decimal.Decimal
}

type PercentageDiscount struct {
    Discount decimal.Decimal
}

type FixedDiscount struct {
    Amount decimal.Decimal
}

func (p *PercentageDiscount) Calculate(items []Item) decimal.Decimal {
    // Calculate percentage discount
}
```

### 4. Decorator Pattern
```go
type LoggingMiddleware struct {
    next   Service
    logger *zap.Logger
}

func (m *LoggingMiddleware) CreateClient(ctx context.Context, req *CreateClientRequest) (*Client, error) {
    m.logger.Info("creating client", zap.String("name", req.Name))
    start := time.Now()
    
    client, err := m.next.CreateClient(ctx, req)
    
    m.logger.Info("client created",
        zap.String("id", client.ID.String()),
        zap.Duration("latency", time.Since(start)),
    )
    return client, err
}
```

### 5. Saga Pattern (for distributed transactions)
```go
type PaymentSaga struct {
    steps       []SagaStep
    compensation []CompensationAction
}

type SagaStep func(ctx context.Context) error

func (s *PaymentSaga) Execute(ctx context.Context) error {
    for _, step := range s.steps {
        if err := step(ctx); err != nil {
            s.compensate(ctx)
            return err
        }
    }
    return nil
}
```

### 6. Observer Pattern (for event handling)
```go
type EventPublisher interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(subject string, handler EventHandler) error
}

// Used for domain events
type Client struct {
    events []Event
}

func (c *Client) Record(event Event) {
    c.events = append(c.events, event)
}
```

### 7. Builder Pattern (for complex objects)
```go
type InvoiceBuilder struct {
    invoice *Invoice
}

func NewInvoiceBuilder() *InvoiceBuilder {
    return &InvoiceBuilder{invoice: &Invoice{}}
}

func (b *InvoiceBuilder) WithClient(clientID uuid.UUID) *InvoiceBuilder {
    b.invoice.ClientID = clientID
    return b
}

func (b *InvoiceBuilder) WithLine(item LineItem) *InvoiceBuilder {
    b.invoice.LineItems = append(b.invoice.LineItems, item)
    return b
}

func (b *InvoiceBuilder) Build() (*Invoice, error) {
    if b.invoice.ClientID == uuid.Nil {
        return nil, ErrClientRequired
    }
    return b.invoice, nil
}
```

---

## Test Implementation Guidelines

### Always Write Tests for New Features

When implementing any new feature, you MUST create tests. This is mandatory:

1. **Domain Models** (`internal/domain/`):
   - Create `*_test.go` file for each domain entity
   - Test all public methods
   - Test edge cases and error conditions
   - Target: 100% coverage (minimum 95%)

2. **Command Handlers** (`internal/commands/`):
   - Test command validation
   - Test business rules enforcement
   - Test event emission
   - Test error handling

3. **Query Handlers** (`internal/queries/`):
   - Test query filtering
   - Test pagination
   - Test projection accuracy

4. **Event Handlers** (`internal/events/`):
   - Test event projection
   - Test read model updates

5. **Services** (`cmd/*/main.go`):
   - Test HTTP handlers (use httptest)
   - Test middleware
   - Test integration points

### Test Coverage Requirements

```bash
# Check coverage for a package
go test -cover ./internal/domain/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Target: 100% for domain, commands, and queries (minimum 95%)
```

### Test Patterns to Follow

**Domain Model Tests:**
```go
func TestClient_Update(t *testing.T) {
    client := NewClient(...)
    
    err := client.Update("new name")
    
    require.NoError(t, err)
    assert.Equal(t, "new name", client.Name)
}

func TestClient_Validation(t *testing.T) {
    tests := []struct {
        name    string
        client  *Client
        wantErr bool
    }{
        {"valid client", validClient, false},
        {"empty name", clientWithEmptyName, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.client.Validate()
            assert.Equal(t, tt.wantErr, err != nil)
        })
    }
}
```

**Command Handler Tests:**
```go
func TestCreateClientCommand_Handler(t *testing.T) {
    handler := NewCreateClientHandler(repo)
    
    cmd := &CreateClientCommand{
        Name:  "Test Client",
        Email: "test@example.com",
    }
    
    err := handler.Handle(context.Background(), cmd)
    
    require.NoError(t, err)
    assert.Len(t, repo.events, 1)
}
```

---

## Achieving 100% Test Coverage

### Coverage Targets

| Package | Minimum | Target | Stretch Goal |
|---------|---------|--------|--------------|
| `internal/domain/` | 95% | 98% | 100% |
| `internal/commands/` | 90% | 95% | 100% |
| `internal/queries/` | 90% | 95% | 100% |
| `internal/events/` | 85% | 90% | 95% |
| `cmd/*/` | 70% | 80% | 90% |

### Coverage Commands

```bash
# Run coverage for specific package
go test -cover -covermode=count ./internal/domain/...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Show uncovered lines
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser

# Coverage by function
go test -cover -covermode=count ./... | grep -E "^(ok|FAIL)"

# Watch coverage over time
while true; do clear; go test -cover ./... 2>/dev/null | tail -5; sleep 5; done
```

### Coverage Strategy for 100%

#### 1. Test All Export Functions

Every exported function MUST have at least one test:

```go
// ✅ GOOD: All exported functions tested
func TestWarehouse_Activate(t *testing.T) { ... }
func TestWarehouse_Deactivate(t *testing.T) { ... }
func TestWarehouse_SetAddress(t *testing.T) { ... }

// ❌ BAD: Missing test for exported function
func TestWarehouse_Activate(t *testing.T) { ... }
// Missing: Deactivate, SetAddress
```

#### 2. Test All Error Paths

Every error condition MUST be tested:

```go
func TestCreateWarehouse_ValidationErrors(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateWarehouse
        wantErr error
    }{
        {
            name: "empty code returns error",
            input: CreateWarehouse{Name: "Test"},
            wantErr: ErrWarehouseCodeRequired,
        },
        {
            name: "empty name returns error",
            input: CreateWarehouse{Code: "WH001"},
            wantErr: ErrWarehouseNameRequired,
        },
        {
            name: "invalid type returns error",
            input: CreateWarehouse{
                Name: "Test",
                Code: "WH001",
                Type: "invalid",
            },
            wantErr: ErrInvalidWarehouseType,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewWarehouse(...)
            assert.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

#### 3. Test Boundary Conditions

```go
func TestInventoryItem_Reserve_Boundaries(t *testing.T) {
    item := &InventoryItem{
        Quantity:     10,
        AvailableQty: 10,
        ReservedQty:  0,
    }

    tests := []struct {
        name        string
        quantity    int
        wantErr     bool
        wantReserve int
        wantAvail   int
    }{
        {"exact quantity", 10, false, 10, 0},
        {"partial quantity", 5, false, 5, 5},
        {"over quantity", 11, true, 0, 10},
        {"zero quantity", 0, false, 0, 10},
        {"negative quantity", -5, true, 0, 10},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := item.Reserve(tt.quantity)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantReserve, item.ReservedQty)
                assert.Equal(t, tt.wantAvail, item.AvailableQty)
            }
        })
    }
}
```

#### 4. Test State Transitions

```go
func TestWarehouseOperation_Lifecycle(t *testing.T) {
    op, err := NewWarehouseOperation(
        tenantID, warehouseID, userID,
        OperationTypeReceipt, "order", orderID,
    )
    require.NoError(t, err)

    assert.Equal(t, "pending", op.Status)
    assert.Nil(t, op.StartedAt)

    op.Start()
    assert.Equal(t, "in_progress", op.Status)
    assert.NotNil(t, op.StartedAt)
    assert.False(t, op.IsComplete())

    // Complete items
    for i := range op.Items {
        err := op.CompleteItem(op.Items[i].ID, op.Items[i].Quantity)
        require.NoError(t, err)
    }
    assert.True(t, op.IsComplete())

    op.Complete()
    assert.Equal(t, "completed", op.Status)
    assert.NotNil(t, op.CompletedAt)
}
```

#### 5. Use Table-Driven Tests for Combinatorial Coverage

```go
func TestReservation_VariousStatuses(t *testing.T) {
    reservation := &StockReservation{
        Status: "active",
    }

    tests := []struct {
        name     string
        call     func(*StockReservation)
        wantStat string
    }{
        {"expire sets expired", func(r *StockReservation) { r.Expire() }, "expired"},
        {"release sets released", func(r *StockReservation) { r.Release() }, "released"},
        {"fulfill sets fulfilled", func(r *StockReservation) { r.Fulfill() }, "fulfilled"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.call(reservation)
            assert.Equal(t, tt.wantStat, reservation.Status)
        })
    }
}
```

#### 6. Test All Repository Methods

```go
func TestWarehouseRepository_CRUD(t *testing.T) {
    repo := NewMongoWarehouseRepository(db)
    ctx := context.Background()
    tenantID := uuid.New()

    // Create
    warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH001", domain.WarehouseTypeMain)
    err := repo.Create(ctx, warehouse)
    require.NoError(t, err)

    // Read
    fetched, err := repo.FindByID(ctx, warehouse.ID)
    require.NoError(t, err)
    assert.Equal(t, warehouse.Name, fetched.Name)

    // Update
    fetched.Name = "Updated WH"
    err = repo.Update(ctx, fetched)
    require.NoError(t, err)

    updated, _ := repo.FindByID(ctx, warehouse.ID)
    assert.Equal(t, "Updated WH", updated.Name)

    // Delete
    err = repo.Delete(ctx, warehouse.ID)
    require.NoError(t, err)

    _, err = repo.FindByID(ctx, warehouse.ID)
    assert.Error(t, err)
}
```

#### 7. Mock External Dependencies

```go
type MockPublisher struct {
    events []events.EventEnvelope
    mu     sync.Mutex
}

func (m *MockPublisher) PublishEvent(ctx context.Context, event *events.EventEnvelope) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.events = append(m.events, *event)
    return nil
}

func (m *MockPublisher) GetEvents() []events.EventEnvelope {
    m.mu.Lock()
    defer m.mu.Unlock()
    return append([]events.EventEnvelope{}, m.events...)
}

func TestCreateWarehouse_PublishesEvent(t *testing.T) {
    publisher := &MockPublisher{}
    handler := NewWarehouseCommandHandler(repo, locationRepo, opRepo, publisher)

    cmd := NewCommand("createWarehouse", tenantID.String(), "", userID, map[string]interface{}{...})
    _, err := handler.HandleCreateWarehouse(context.Background(), cmd)

    require.NoError(t, err)
    events := publisher.GetEvents()
    assert.Len(t, events, 1)
    assert.Equal(t, "warehouse.created", events[0].Type)
}
```

#### 8. Coverage Gaps Analysis

When coverage is below 100%, use this process:

```bash
# 1. Generate coverage report
go test -coverprofile=coverage.out ./internal/domain/...

# 2. Find uncovered lines
go tool cover -func=coverage.out | grep "0.0%"

# 3. Add tests for each gap
# 4. Re-run coverage
# 5. Repeat until 100%
```

Common gaps to watch:
- `else` branches
- `switch` cases
- Error conditions
- Nil pointer checks
- Boundary conditions
- Default cases in type switches

#### 9. Coverage Quality Checks

100% coverage doesn't mean good tests. Verify:

```go
// ✅ GOOD: Tests actual behavior
func TestWarehouse_Deactivate_WithActiveStock_ReturnsError(t *testing.T) {
    warehouse := NewWarehouse(...)
    warehouse.AddLocation(locationWithStock)
    
    err := warehouse.Deactivate()
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "stock")
}

// ❌ BAD: 100% coverage but no real test
func TestWarehouse_Deactivate(t *testing.T) {
    warehouse := NewWarehouse(...)
    warehouse.Deactivate()  // What are we testing?
}
```

### CI/CD Coverage Enforcement

Add to your CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: |
    go test -coverprofile=coverage.out ./...
    
    # Check domain package
    go tool cover -func=coverage.out | grep "internal/domain"
    go tool cover -func=coverage.out | awk '/internal\/domain/ {gsub(/%|[[:space:]]/,"",$3); if($3+0 < 100) exit 1}'
    
    # Upload coverage to codecov
    curl -s https://codecov.io/bash | bash -s -- -f coverage.out
```

---

## Plan Progress Tracking

### Files to Update

When completing tasks, update these files:

1. **`AGENTS.md`** - Update todo list at the end
2. **`IMPLEMENTATION-PLAN.md`** - Mark tasks as complete in "Success Criteria" section
3. **`README.md`** - Update features list if needed

### How to Update AGENTS.md Todo List

Use the `todowrite` tool to track progress:

```bash
# At start of session, read current todos
todowrite --todos []

# Mark tasks as in_progress
todowrite --todos [{"id": "1", "content": "Implement client module", "status": "in_progress"}]

# Mark tasks as completed
todowrite --todos [{"content": "Implement client module", "id": "1", "status": "completed"}]
```

### How to Update IMPLEMENTATION-PLAN.md

After completing a task:

1. Find the relevant section in "Success Criteria"
2. Change `[ ]` to `[x]` for completed items
3. Add notes about what was implemented
4. Update the "Implementation Status" table

Example:
```markdown
### Phase 1 - Foundation
- [x] Client module fully functional with CQRS ✅ **Domain model + 95 tests passing**
- [ ] Authentication and RBAC working
```

### Progress Reporting Format

When reporting completed work, use this format:

```markdown
## Completed Tasks [Date]

| Task | Status | Notes |
|------|--------|-------|
| Create client domain model | ✅ Complete | 23 tests passing |
| Add JWT authentication | ⚠️ Partial | OAuth not implemented |
| Fix payment handler bug | ✅ Complete | Edge case handled |

**Test Results:** `go test ./...` - 95 passing, 0 failing
**Build Status:** `go build ./...` - All packages compile
```

### Session Summary Template

When finishing a session, output:

```markdown
## Session Summary

### Completed
- [x] Task 1
- [x] Task 2

### Test Results
- 95 unit tests passing
- 0 integration tests (require infrastructure)

### Files Modified
- `internal/domain/client.go`
- `internal/domain/client_test.go`
- `IMPLEMENTATION-PLAN.md`

### Next Steps
1. Continue with Phase 1 tasks
2. Implement OAuth integration
3. Add payment processor stubs
```

### Quick Reference: Update Commands

```bash
# Run tests and verify before marking complete
go test ./... -short

# Verify build succeeds
go build ./...

# Update plan status (in IMPLEMENTATION-PLAN.md)
# Change: - [ ] Task Name
# To:     - [x] Task Name ✅ **Completed**

# Update todo list
todowrite --todos [...]
```

---

## Code Refactoring Guidelines

### Prefer Refactoring Over Leaving Legacy Code

When working in the codebase, if you encounter code that doesn't match the guidelines in this file, **refactor it to conform**. Don't leave legacy patterns scattered throughout the codebase.

### When to Refactor

**Refactor legacy code when you:**
- Fix bugs in existing code
- Add new features to existing code
- Modify code during debugging
- Work on adjacent code in the same file

**Examples of code that needs refactoring:**

```go
// ❌ AVOID: Legacy patterns that don't match AGENTS.md guidelines

// 1. Mutex instead of channels for complex state
type Service struct {
    mu sync.Mutex
    data map[string]string
}

func (s *Service) Get(key string) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.data[key]
}

// 2. Inconsistent naming (user_id instead of userID)
type Order struct {
    user_id string
    order_id string
}

// 3. Poor error handling
func (s *Service) DoSomething() error {
    err := doThing()
    if err != nil {
        return err  // No context
    }
    return nil
}

// 4. No documentation
func process(data []byte) []byte {
    // What does this do?
    return transform(data)
}

// 5. No tests
func calculateTotal(items []Item) decimal.Decimal {
    // Business logic without tests!
    return total
}

// ✅ REFACTORED: Follow AGENTS.md guidelines

type Service struct {
    data     chan map[string]string  // Use channels
    shutdown chan struct{}
}

func NewService() *Service {
    return &Service{
        data:     make(chan map[string]string, 100),
        shutdown: make(chan struct{}),
    }
}

// 1. Consistent naming
type Order struct {
    UserID  string
    OrderID string
}

// 2. Context-rich errors
func (s *Service) DoSomething(ctx context.Context) error {
    if err := doThing(); err != nil {
        return fmt.Errorf("failed to do thing: %w", err)
    }
    return nil
}

// 3. Documentation
// processData transforms raw input into processed output.
func processData(data []byte) ([]byte, error) {
    return transform(data)
}

// 4. Tests required
func TestCalculateTotal(t *testing.T) {
    items := []Item{{Price: decimal.NewFromFloat(10)}}
    total := calculateTotal(items)
    assert.True(t, total.Equal(decimal.NewFromFloat(10)))
}
```

### Refactoring Priority

**High Priority (always refactor):**
- Mutex usage that should be channels
- Missing documentation on public APIs
- Poor error handling (missing context)
- Inconsistent naming conventions

**Medium Priority (refactor when touching):**
- Missing tests (add when modifying)
- Complex functions (simplify when modifying)
- Missing logging (add when modifying)

**Low Priority (refactor when significant work):**
- Large struct reorganization
- Package structure changes
- Pattern migrations

### Refactoring Workflow

1. **Identify** - Find code not following guidelines
2. **Assess** - Determine refactoring priority
3. **Plan** - Plan changes before coding
4. **Refactor** - Make incremental changes
5. **Test** - Ensure tests still pass
6. **Verify** - Run `go test -short ./...` and `go build ./...`

### Don't Leave Technical Debt

> "The best time to refactor legacy code is when you're already working on it. The second best time is now."

When you see violations of AGENTS.md guidelines:
- Fix them immediately if quick
- Add TODO comment if significant work: `// TODO: Refactor to use channels instead of mutex`
- Track in task list for future sprint

### Example TODO Tracking

```go
// TODO: Refactor to use channels for state sharing
// See AGENTS.md "Prefer Channels Over Mutexes" section
type LegacyService struct {
    mu sync.Mutex  // Should be channels
    // ...
}
```

### Quick Refactor Checklist

When modifying existing code, verify:

- [ ] Naming follows conventions (camelCase vars, PascalCase types)
- [ ] Error handling includes context with `fmt.Errorf(...: %w, err)`
- [ ] Context passed as first parameter
- [ ] Documentation on public functions
- [ ] Tests added/updated for changed logic
- [ ] Channels used instead of mutexes for goroutine communication
- [ ] Design patterns applied where appropriate
- [ ] Logging added for important operations

---

## Frontend Development Guidelines

### Frontend Architecture Reference

When working on the frontend, always reference:

1. **`frontend/ARCHITECTURE.md`** - SvelteKit plugin system architecture
2. **`frontend/UI-DESIGN-GUIDE.md`** - Professional UI design standards
3. **`frontend/PLAN.md`** - Frontend implementation roadmap

### Frontend Tech Stack

| Technology | Purpose |
|------------|---------|
| SvelteKit | Application framework |
| Tailwind CSS | Utility-first styling |
| Lucide Svelte | Icon library |
| Zod | Form validation |
| date-fns | Date formatting |

### Frontend Commands

```bash
# Navigate to frontend directory
cd frontend/

# Install dependencies
npm install

# Install additional UI dependencies
npm install -D tailwindcss postcss autoprefixer
npm install lucide-svelte clsx tailwind-merge zod date-fns

# Run development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint TypeScript
npm run lint

# Type check
npm run typecheck
```

### Frontend Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── core/                    # Plugin system core
│   │   │   ├── index.ts
│   │   │   ├── types.ts
│   │   │   ├── plugin-registry.ts
│   │   │   ├── plugin-loader.ts
│   │   │   ├── message-bus.ts
│   │   │   ├── route-manager.ts
│   │   │   ├── state-manager.ts
│   │   │   └── permissions.ts
│   │   │
│   │   ├── shared/                  # Shared utilities
│   │   │   ├── api/                 # API client layer
│   │   │   ├── components/          # Reusable UI components
│   │   │   │   ├── display/         # Badge, Card, Alert, Toast, etc.
│   │   │   │   ├── forms/           # Input, Button, Select, etc.
│   │   │   │   ├── layout/          # Sidebar, Modal, Card, etc.
│   │   │   │   └── data/            # Table, Pagination, etc.
│   │   │   ├── utils/               # Formatting, validation helpers
│   │   │   ├── styles/              # CSS variables, global styles
│   │   │   └── types/               # TypeScript types
│   │   │
│   │   └── plugins/                 # Feature plugins
│   │       ├── dashboard/
│   │       ├── clients/
│   │       ├── warehouse/
│   │       ├── inventory/
│   │       ├── products/
│   │       ├── users/
│   │       ├── documents/
│   │       ├── invoices/
│   │       ├── payments/
│   │       └── orders/
│   │
│   └── routes/
│       ├── +layout.svelte           # Root layout
│       ├── +layout.server.ts        # Server-side plugin loading
│       ├── +page.svelte             # Main app shell
│       └── [[...catchall]]/         # Dynamic plugin routes
```

### Frontend Code Style

#### Component Structure

```svelte
<!-- src/lib/shared/components/forms/Button.svelte -->
<script lang="ts">
  // 1. Props with types
  export let variant: 'primary' | 'secondary' | 'danger' = 'primary';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let disabled = false;
  export let loading = false;

  // 2. Event handlers
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function handleClick() {
    if (!disabled && !loading) {
      dispatch('click');
    }
  }
</script>

<!-- 3. Component markup with Tailwind classes -->
<button
  class="inline-flex items-center justify-center font-medium rounded-lg transition-colors"
  class:bg-primary-600={variant === 'primary'}
  class:hover:bg-primary-700={variant === 'primary'}
  class:bg-gray-100={variant === 'secondary'}
  class:opacity-50={disabled || loading}
  {disabled}
  on:click={handleClick}
>
  {#if loading}
    <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  {/if}
  <slot />
</button>
```

#### API Client Pattern

```typescript
// src/lib/shared/api/clients.ts
import type { Client, CreateClientRequest, ClientFilter } from '$lib/shared/types';

const API_BASE = '/api/clients';

export async function getClients(filter?: ClientFilter): Promise<Client[]> {
  const params = new URLSearchParams();
  if (filter?.status) params.set('status', filter.status);
  if (filter?.page) params.set('page', filter.page.toString());
  
  const response = await fetch(`${API_BASE}?${params}`);
  if (!response.ok) throw new Error('Failed to fetch clients');
  return response.json();
}

export async function getClientById(id: string): Promise<Client> {
  const response = await fetch(`${API_BASE}/${id}`);
  if (!response.ok) throw new Error('Client not found');
  return response.json();
}

export async function createClient(data: CreateClientRequest): Promise<Client> {
  const response = await fetch(API_BASE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error('Failed to create client');
  return response.json();
}

export async function updateClient(id: string, data: Partial<CreateClientRequest>): Promise<Client> {
  const response = await fetch(`${API_BASE}/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error('Failed to update client');
  return response.json();
}

export async function deleteClient(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/${id}`, { method: 'DELETE' });
  if (!response.ok) throw new Error('Failed to delete client');
}
```

#### Plugin Definition Pattern

```typescript
// src/lib/plugins/clients/index.ts
import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';
import { api } from './api';
import { messages } from './messages';
import { stores } from './stores';
import { routes } from './routes';

const clientsPlugin: PluginDefinition = {
  manifest,
  api,
  messages,
  stores,
  routes,

  async setup(context) {
    context.logger.info('Clients plugin initializing...');
    // Initialize plugin state
  },

  async teardown() {
    // Cleanup
  },
};

export default clientsPlugin;
```

### Frontend Testing

```bash
# Run frontend tests
cd frontend && npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run e2e tests
npm run test:e2e
```

### E2E Testing Guidelines

E2E tests verify the complete application flow from user perspective.

#### E2E Testing Stack

| Tool | Purpose |
|------|---------|
| Playwright | Browser automation |
| Vitest | Test runner |
| Cucumber | BDD scenarios |

#### E2E Test Commands

```bash
# Install Playwright
npm install -D @playwright/test
npx playwright install

# Run all e2e tests
npm run test:e2e

# Run specific test file
npm run test:e2e -- tests/clients.spec.ts

# Run tests in headed mode (see browser)
npm run test:e2e -- --headed

# Run tests with debugging
npm run test:e2e -- --debug

# Generate tests with Codegen
npx playwright codegen http://localhost:5173

# Take screenshot on failure
npm run test:e2e -- --reporter=line --grep="@smoke"

# Run tests in parallel
npm run test:e2e -- --workers=4
```

#### E2E Test Structure

```
frontend/
├── tests/
│   ├── e2e/
│   │   ├── setup/
│   │   │   ├── fixtures.ts          # Test fixtures
│   │   │   ├── authentication.ts    # Login helper
│   │   │   └── api-helpers.ts       # API setup helpers
│   │   ├── pages/
│   │   │   ├── BasePage.ts          # Base page object
│   │   │   ├── DashboardPage.ts
│   │   │   ├── ClientsPage.ts
│   │   │   ├── WarehousesPage.ts
│   │   │   └── InventoryPage.ts
│   │   ├── specs/
│   │   │   ├── dashboard.spec.ts    # Dashboard tests
│   │   │   ├── clients.spec.ts      # Client management
│   │   │   ├── warehouses.spec.ts   # Warehouse management
│   │   │   ├── inventory.spec.ts    # Inventory tests
│   │   │   ├── products.spec.ts     # Product tests
│   │   │   ├── users.spec.ts        # User management
│   │   │   ├── documents.spec.ts    # Document tests
│   │   │   ├── invoices.spec.ts     # Invoice tests
│   │   │   ├── payments.spec.ts     # Payment tests
│   │   │   └── orders.spec.ts       # Order tests
│   │   ├── helpers/
│   │   │   ├── selectors.ts         # Reusable selectors
│   │   │   └── constants.ts         # Test constants
│   │   └── playwright.config.ts
│   └── integration/
│       └── ...                      # Integration tests
```

#### E2E Test Example

```typescript
// tests/e2e/specs/clients.spec.ts
import { test, expect } from '@playwright/test';
import { ClientsPage } from '../pages/ClientsPage';

test.describe('Client Management', () => {
  let clientsPage: ClientsPage;

  test.beforeEach(async ({ page }) => {
    clientsPage = new ClientsPage(page);
    await clientsPage.goto();
  });

  test('should display client list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Clients');
    await expect(page.locator('[data-testid="client-list"]')).toBeVisible();
  });

  test('should create a new client', async ({ page }) => {
    await page.click('[data-testid="create-client-button"]');
    await expect(page.locator('[data-testid="client-form"]')).toBeVisible();

    await page.fill('[data-testid="name-input"]', 'Test Company');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="phone-input"]', '+1234567890');

    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText(
      'Client created successfully'
    );
    await expect(page.locator('[data-testid="client-name"]')).toHaveText('Test Company');
  });

  test('should search clients', async ({ page }) => {
    await page.fill('[data-testid="search-input"]', 'Test Company');
    await page.click('[data-testid="search-button"]');

    await expect(page.locator('[data-testid="client-row"]').first()).toContainText('Test Company');
  });

  test('should edit client details', async ({ page }) => {
    await page.click('[data-testid="client-row"] >> nth=0');
    await expect(page).toHaveURL(/\/clients\/.+/);

    await page.click('[data-testid="edit-button"]');
    await page.fill('[data-testid="name-input"]', 'Updated Company');

    await page.click('[data-testid="save-button"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText('Client updated');
  });

  test('should show validation errors', async ({ page }) => {
    await page.click('[data-testid="create-client-button"]');
    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('[data-testid="name-error"]')).toContainText('Name is required');
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Email is required');
  });

  test('should delete client', async ({ page }) => {
    await page.hover('[data-testid="client-row"] >> nth=0');
    await page.click('[data-testid="delete-button"]');

    await expect(page.locator('[data-testid="confirm-dialog"]')).toBeVisible();
    await page.click('[data-testid="confirm-delete"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText('Client deleted');
  });
});

test.describe('Client Search and Filter', () => {
  test('should filter by status', async ({ page }) => {
    await page.selectOption('[data-testid="status-filter"]', 'active');
    await page.click('[data-testid="apply-filters"]');

    await expect(page.locator('[data-testid="client-row"]')).toHaveCount(5);
  });

  test('should clear filters', async ({ page }) => {
    await page.selectOption('[data-testid="status-filter"]', 'active');
    await page.click('[data-testid="clear-filters"]');

    await expect(page.locator('[data-testid="status-filter"]')).toHaveValue('');
  });
});
```

#### Page Object Pattern

```typescript
// tests/e2e/pages/ClientsPage.ts
import type { Page } from '@playwright/test';

export class ClientsPage {
  constructor(private page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto('/clients');
  }

  async createClient(data: {
    name: string;
    email: string;
    phone?: string;
  }): Promise<void> {
    await this.page.click('[data-testid="create-client-button"]');
    await this.page.fill('[data-testid="name-input"]', data.name);
    await this.page.fill('[data-testid="email-input"]', data.email);
    if (data.phone) {
      await this.page.fill('[data-testid="phone-input"]', data.phone);
    }
    await this.page.click('[data-testid="submit-button"]');
  }

  async searchClients(query: string): Promise<void> {
    await this.page.fill('[data-testid="search-input"]', query);
    await this.page.click('[data-testid="search-button"]');
  }

  async getClientNames(): Promise<string[]> {
    return this.page.locator('[data-testid="client-name"]').allTextContents();
  }

  async getClientCount(): Promise<number> {
    return this.page.locator('[data-testid="client-row"]').count();
  }
}
```

#### Test Fixtures

```typescript
// tests/e2e/setup/fixtures.ts
import { test as base } from '@playwright/test';
import { ClientsPage } from '../pages/ClientsPage';
import { WarehousesPage } from '../pages/WarehousesPage';

interface Fixtures {
  clientsPage: ClientsPage;
  warehousesPage: WarehousesPage;
  authenticatedPage: Page;
}

export const test = base.extend<Fixtures>({
  authenticatedPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    // Login before each test
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@example.com');
    await page.fill('[data-testid="password"]', 'password123');
    await page.click('[data-testid="login-button"]');

    await expect(page).toHaveURL('/dashboard');

    await use(page);

    await context.close();
  },

  clientsPage: async ({ authenticatedPage }, use) => {
    const clientsPage = new ClientsPage(authenticatedPage);
    await use(clientsPage);
  },

  warehousesPage: async ({ authenticatedPage }, use) => {
    const warehousesPage = new WarehousesPage(authenticatedPage);
    await use(warehousesPage);
  },
});
```

#### Authentication Fixture

```typescript
// tests/e2e/setup/authentication.ts
import type { Page } from '@playwright/test';

export async function loginAs(page: Page, role: 'admin' | 'user' | 'viewer'): Promise<void> {
  const credentials = {
    admin: { email: 'admin@example.com', password: 'password123' },
    user: { email: 'user@example.com', password: 'password123' },
    viewer: { email: 'viewer@example.com', password: 'password123' },
  };

  await page.goto('/login');
  await page.fill('[data-testid="email"]', credentials[role].email);
  await page.fill('[data-testid="password"]', credentials[role].password);
  await page.click('[data-testid="login-button"]');
}

export async function logout(page: Page): Promise<void> {
  await page.click('[data-testid="user-menu"]');
  await page.click('[data-testid="logout-button"]');
}
```

#### E2E Test Checklist

Before running E2E tests:

- [ ] Backend services are running
- [ ] Frontend dev server is running (`npm run dev`)
- [ ] Test database is available
- [ ] No sensitive data in tests

Before committing E2E tests:

- [ ] Tests pass locally
- [ ] Tests are independent (can run in any order)
- [ ] Tests handle loading states
- [ ] Tests use data-testid selectors
- [ ] Tests include error scenarios
- [ ] Tests use page objects for complex flows
- [ ] Tests are readable and well-documented
- [ ] No hardcoded waits (use expect or waitFor)

#### E2E Test Coverage Targets

| Feature | Minimum Tests | Description |
|---------|--------------|-------------|
| Authentication | 5 | Login, logout, session expiry, MFA, permissions |
| Dashboard | 3 | Load, widgets, navigation |
| Clients | 10 | CRUD, search, filter, validation, errors |
| Warehouses | 8 | CRUD, locations, operations |
| Inventory | 8 | Items, reservations, transactions |
| Products | 8 | CRUD, variants, pricing |
| Users | 6 | CRUD, roles, permissions |
| Documents | 5 | Upload, search, download |
| Invoices | 7 | CRUD, line items, payments |
| Payments | 5 | Record, refund, reconciliation |
| Orders | 7 | CRUD, fulfillment, tracking |

### Frontend Linting

```bash
# Lint and format
cd frontend && npm run lint
npm run format

# Type check
npm run typecheck
```

### UI Design Checklist

Before committing frontend code, verify:

- [ ] Components use shared UI library (Button, Input, Card, etc.)
- [ ] Colors follow the design system (primary, neutral, semantic)
- [ ] Typography uses the type scale (text-xs to text-4xl)
- [ ] Spacing follows the 4px base unit (space-1 to space-20)
- [ ] Components have hover, focus, and disabled states
- [ ] Forms include validation and error messages
- [ ] Loading states use skeletons or spinners
- [ ] Dark mode is supported
- [ ] Responsive design works on mobile
- [ ] Accessibility: ARIA labels, keyboard navigation, focus indicators

### Frontend Resources

| Resource | Location |
|----------|----------|
| Plugin System | `frontend/ARCHITECTURE.md` |
| UI Design Guide | `frontend/UI-DESIGN-GUIDE.md` |
| Implementation Plan | `frontend/PLAN.md` |
| Component Examples | `frontend/src/lib/shared/components/` |
| Plugin Examples | `frontend/src/lib/plugins/dashboard/` |


