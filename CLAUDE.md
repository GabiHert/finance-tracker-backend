# Golang Backend Development Guidelines - Clean Architecture

## Mandatory Agent Usage

**CRITICAL RULE**: All code, scripts, configuration files, and documentation in this repository MUST be created, modified, or reviewed by the `@golang-clean-arch-specialist` agent.
**CRITICAL RULE**: Do not mistake the specialist with the reviewer agent, you must invoke the `@golang-clean-arch-specialist` for any work in this repository not the `@golang-clean-arch-reviewer` agent.

### When to Use the Agent

The agent MUST be used for:

1. **Creating or modifying any files** in this repository
2. **Writing scripts** (bash, deployment, orchestration)
3. **Creating Docker configurations** (Dockerfiles, docker-compose files)
4. **Updating documentation**
5. **Designing system architecture**
6. **Troubleshooting deployment issues**
7. **Standardizing service repositories**
8. **Setting up infrastructure** (databases, Redis, LocalStack)
9. **Creating health checks and monitoring**
10. **Configuring environment variables**

### How to Invoke the Agent

When a user requests work on this repository, immediately invoke:

```
@golang-clean-arch-specialist
```

Or programmatically:

```
Use Task tool with subagent_type="golang-clean-arch-specialist"
```

### Exception: User Explicitly Opts Out

The ONLY exception is if the user explicitly states they want to work without the agent. In this case, confirm with the user first before proceeding.

### Enforcement

- Do NOT write code directly without using the agent
- Do NOT create files without agent involvement
- Do NOT make architectural decisions without agent consultation
- Always route tasks through the agent

This ensures consistency, best practices, and adherence to the established patterns in this repository.

## General Guidelines

### Development Philosophy
- Write idiomatic, maintainable, and scalable Go code
- Follow Clean Architecture principles with clear separation of concerns
- Emphasize SOLID principles and Domain-Driven Design (DDD)
- Prefer composition over inheritance
- Design for testability with dependency injection
- Handle errors explicitly - never ignore them
- Make zero values useful
- Keep interfaces small and focused
- Avoid premature optimization
- Every package should have a single, well-defined purpose
- Prioritize simplicity and readability
- Eliminate technical debt through continuous refactoring

### Architecture Principles
- **Independence of Frameworks**: Business logic should not depend on external frameworks
- **Testability**: Business rules can be tested without UI, database, or external services
- **Independence of UI**: Business rules don't know about the delivery mechanism
- **Independence of Database**: Business rules are not bound to database logic
- **Independence of External Services**: Business rules don't know about the outside world

## Folder Structure

### Root Structure
```
/
├── .github/              # GitHub Actions workflows and CI/CD configuration
│   └── workflows/        # CI/CD pipeline definitions
├── build/                # Docker and build configurations
│   ├── docker/          # Dockerfiles for different environments
│   └── compose/         # Docker Compose configurations
├── claude/              # AI prompts and documentation
├── cloudformation/      # AWS CloudFormation infrastructure
│   ├── parameters/      # Environment-specific parameters
│   │   ├── dev/
│   │   ├── staging/
│   │   └── prod/
│   └── templates/       # CloudFormation template files
├── cmd/                 # Application entry points
│   └── <app-name>/     # Main application
│       └── main.go
├── internal/            # Private application code (not importable by external packages)
│   ├── application/     # Application business logic layer
│   ├── domain/         # Core business logic and entities
│   ├── infra/          # Infrastructure and configuration
│   └── integration/    # External service integrations
├── pkg/                # Public packages (importable by external applications)
│   ├── logger/         # Shared logging utilities
│   ├── errors/         # Custom error types and handling
│   └── validator/      # Shared validation utilities
├── scripts/            # Build, deployment, and utility scripts
│   ├── migrations/     # Database migration scripts
│   └── setup/          # Environment setup scripts
├── test/               # Test files and utilities
│   ├── unit/          # Unit tests
│   ├── integration/   # Integration tests
│   ├── e2e/           # End-to-end tests
│   └── fixtures/      # Test data and mocks
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
├── Makefile           # Build and development tasks
└── README.md          # Project documentation
```

### Layer Structure Details

#### `/internal/application/`
Application services and use cases - orchestrates the flow of data between external world and domain.
```
/internal/application/
├── adapter/           # Interfaces the services are allowed to use (these will be implemented in the integration layer)
├── service/           # Application services (multiple entity related actions)
├── usecase/           # Application use cases (singular entity related actions)
└── factories/         # Factories for creating application services
```

#### `/internal/domain/`
Core business logic - the heart of the application, completely independent of external concerns.
```
/internal/domain/
├── entity/             # Business entities
│   ├── user.go
├── enums/              # Enums
│   ├── role/value.go   # (e.g. Role enum values)
│   ├── role.go         # (e.g. Role enum type)
└── error/              # Domain-specific errors
    └── client_already_exists.go
```

#### `/internal/infra/`
Infrastructure concerns - frameworks, databases, configuration.
```
/internal/infra/
├── db/                 # Database connection configuration
│   ├── redis.go
│   └── postgresql.go
├── server/            # HTTP/gRPC server setup
│   ├── adapter/
│   │   └── query_validator.go # Adapters used by the router
│   ├── context/
│   │   └── context.go          # Context proxy implementation for fiber that implements context from integration and is used by controllers
│   ├── middleware/
│   │   └── fiber_middlware.go   # Context proxy implementation for fiber that implements middlware from integration and is used by controllers
│   ├── router/
│   │   └── router.go           # Application router
├── dependency/                # Dependency injection container
│   └── intecjtor.go
└── application.go          # Application start 
```

#### `/internal/integration/`
External service integrations and adapters.
```
/internal/integration/
├── adapters/       # Adapters for integration that implement the adapters from application layer (to avoid coupling integraiton code in the application layer)
├── entrypoint/         # Entrypoint for the app
│   ├── controller/         # Controllers (implementation)
│   │   ├── user.go
│   ├── dto/         # Controllers dtos 
│   │   ├── user.go
│   ├── enums/         # Controllers enums 
│   │   ├── user.go
│   ├── error/         # Controllers errors 
│   │   ├── email_missing.go
│   │   ├── header_missing.go
│   ├── middleware/         # Controllers middlewares to validate requests
│   │   ├── authorization.go
│   ├── validator/         # Validator implementation 
│   │   ├── custom.go
├── persistence/      # Repository implementations
│   ├── model/
│   │   └── user.go
│   │   └── role.go
│   │   └── access.go
│   │   └── client.go
│   └── user.go       # Persistence logic
│   └── role.go       # Persistence logic
│   └── client.go       # Persistence logic
├── webservice/           # External API clients
│   ├── dto/
│   │   └── access_token.go  #model for requests
│   ├── credentials         # request implementation
```

## Naming Conventions

### General Go Conventions
- **Packages**: lowercase, single-word names (avoid underscores)
- **Files**: lowercase with underscores (e.g., `user.go`)
- **Exported identifiers**: PascalCase (e.g., `User`)
- **Unexported identifiers**: camelCase (e.g., `user`)
- **Interfaces**: typically end with `-er` suffix (e.g., `Reader`, `UserRepository`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Acronyms**: Keep consistent case (e.g., `HTTPServer`, not `HttpServer`)

### Clean Architecture Naming
- **Use Cases**: verb + noun (e.g., `CreateUser`, `AuthenticateUser`)
- **Repositories**: entity + `Repository` (e.g., `UserRepository`)
- **Services**: domain/feature + `Service` (e.g., `AuthService`)
- **DTOs**: purpose + `DTO` or `Request`/`Response` (e.g., `CreateUserRequest`)
- **Value Objects**: descriptive noun (e.g., `Email`, `Money`)
- **Domain Events**: past tense (e.g., `UserCreated`, `OrderPlaced`)

## Code Implementation Guidelines

### Planning Phase

**MANDATORY STEPS** (in this exact order):

1. **Create BDD Feature File**: Write `.feature` file in `/test/integration/features/` BEFORE any code
2. **Review Existing Code**: Study similar features to understand established patterns
3. **Map to Existing Steps**: Ensure your scenarios use only existing BDD step definitions
4. **Domain Modeling**: Model following existing entity patterns
5. **Design Interfaces**: Match interface patterns from existing code
6. **Document API Contracts**: Use same documentation style as existing APIs
7. **Plan Error Handling**: Follow existing error handling strategies
8. **Consider Concurrency**: Use same concurrency patterns as existing code

### Code Reuse and Consistency

**CRITICAL**: Always check for existing code before creating new implementations:

```go
// BEFORE creating any new function/struct/interface:
// 1. Search for similar existing functionality
// 2. Check if existing code can be extended/reused
// 3. Follow the EXACT same patterns if new code is needed

// Example: Before creating a new validation function
// WRONG - Creating new validation logic
func validateEmail(email string) error {
    // New validation logic
}

// CORRECT - First check if validation already exists
// If it exists in pkg/validator, use it:
import "github.com/yourproject/pkg/validator"
err := validator.ValidateEmail(email)

// If you MUST create new validation, follow existing patterns exactly:
// If existing validations return ValidationError, yours must too
func validatePhone(phone string) ValidationError {
    // Following the SAME pattern as validateEmail
}
```

### Style Consistency Enforcement

```go
// Study existing code style and replicate EXACTLY:

// If existing code uses this comment style:
// UserService handles user-related operations.
type UserService struct {}

// Don't use different style:
/*
 * ProductService handles product operations
 */
type ProductService struct {} // WRONG - different comment style

// If existing code groups imports like this:
import (
    "context"
    "fmt"
    
    "github.com/yourproject/internal/domain"
    "github.com/yourproject/pkg/logger"
    
    "github.com/external/package"
)

// Your code MUST follow the same grouping pattern
```

### Code Style (Following Effective Go)
```go
// Package declaration with documentation
// Package user provides user management functionality.
package user

import (
    "context"
    "fmt"
    
    "github.com/yourdomain/yourapp/internal/domain/entities"
)

// Interface documentation should describe the contract
// UserService defines operations for user management.
type UserService interface {
    // CreateUser creates a new user in the system.
    CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error)
}

// Struct fields should be grouped logically
type userService struct {
    // Dependencies
    userRepo   UserRepository
    emailSvc   EmailService
    
    // Configuration
    maxRetries int
    timeout    time.Duration
}

// Constructor should validate dependencies
func NewUserService(repo UserRepository, email EmailService) (*userService, error) {
    if repo == nil {
        return nil, errors.New("user repository is required")
    }
    if email == nil {
        return nil, errors.New("email service is required")
    }
    
    return &userService{
        userRepo:   repo,
        emailSvc:   email,
        maxRetries: 3,
        timeout:    30 * time.Second,
    }, nil
}
```

### Error Handling
```go
// Define domain errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrEmailExists  = errors.New("email already exists")
)

// Wrap errors with context
func (s *userService) GetUser(ctx context.Context, id string) (*entities.User, error) {
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

// Use custom error types for rich error information
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

### Dependency Injection
```go
// Use interfaces for dependencies
type UserController struct {
    userService application.UserService
    logger      logger.Logger
}

// Wire dependencies explicitly
func NewUserController(svc application.UserService, log logger.Logger) *UserController {
    return &UserController{
        userService: svc,
        logger:      log,
    }
}
```

## Domain-Driven Design

### Entities
```go
// Entities have identity and lifecycle
type User struct {
    ID        string
    Email     valueobjects.Email
    Username  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business logic belongs in entities
func (u *User) ChangeEmail(email valueobjects.Email) error {
    if err := email.Validate(); err != nil {
        return fmt.Errorf("invalid email: %w", err)
    }
    u.Email = email
    u.UpdatedAt = time.Now()
    return nil
}
```

### Value Objects
```go
// Value objects are immutable and validated
type Email struct {
    value string
}

func NewEmail(value string) (Email, error) {
    if !isValidEmail(value) {
        return Email{}, errors.New("invalid email format")
    }
    return Email{value: strings.ToLower(value)}, nil
}

func (e Email) String() string {
    return e.value
}

func (e Email) Equals(other Email) bool {
    return e.value == other.value
}
```

### Aggregates
```go
// Aggregates ensure consistency boundaries
type Order struct {
    ID         string
    CustomerID string
    Items      []OrderItem
    Total      Money
    Status     OrderStatus
}

func (o *Order) AddItem(product Product, quantity int) error {
    if o.Status != OrderStatusDraft {
        return errors.New("cannot modify confirmed order")
    }
    
    item := OrderItem{
        ProductID: product.ID,
        Quantity:  quantity,
        Price:     product.Price,
    }
    
    o.Items = append(o.Items, item)
    o.recalculateTotal()
    return nil
}
```

## Testing

### BDD (Behavior-Driven Development) - MANDATORY

**CRITICAL REQUIREMENT**: Every new feature or modification MUST begin with a `.feature` file before any implementation.

#### Feature Files Location
All feature files MUST be placed in `/test/integration/features/`:
```
/test/integration/features/
├── user-management.feature
├── authentication.feature
├── order-processing.feature
└── payment-handling.feature
```

#### BDD Development Process

1. **Create Feature File FIRST**:
```gherkin
# /test/integration/features/user-registration.feature
Feature: User Registration
  As a new user
  I want to register an account
  So that I can access the system

  Scenario: Successful user registration
    Given I have valid registration details
    When I submit the registration form
    Then I should receive a confirmation email
    And my account should be created

  Scenario: Registration with existing email
    Given an account with email "user@example.com" already exists
    When I try to register with email "user@example.com"
    Then I should see an error "Email already exists"
    And no new account should be created
```

2. **Use ONLY Existing Step Definitions**:
```go
// DO NOT CREATE NEW STEP DEFINITIONS
// The existing generic steps are sufficient for all scenarios
// Example of existing generic steps you should use:

// API steps
Given("I have valid {string} details", ...)
When("I send a {string} request to {string}", ...)
Then("the response status should be {int}", ...)
Then("the response should contain {string}", ...)

// Database steps  
Given("a {string} with {string} {string} already exists", ...)
Then("a {string} should exist with {string} {string}", ...)
Then("no {string} should exist with {string} {string}", ...)

// Generic validation steps
Then("I should see an error {string}", ...)
Then("the {string} field should be {string}", ...)
```

3. **Map Features to Existing Steps**:
```go
// WRONG - Creating new specific steps
func iHaveAUserWithEmail(email string) error {
    // Don't do this!
}

// CORRECT - Use existing generic steps
// The existing "a {string} with {string} {string} already exists" step
// can handle: "a user with email user@example.com already exists"
```

#### Code Consistency Requirements

**MANDATORY**: Before writing ANY new code:

1. **Study Existing Patterns**:
```go
// First, examine how similar features are implemented
// If UserService exists, follow its patterns for new services:

// WRONG - Inventing new patterns
type ProductManager struct {  // Different naming pattern
    db *sql.DB  // Direct DB dependency
}

// CORRECT - Following existing UserService pattern
type ProductService struct {  // Same naming convention
    productRepo ProductRepository  // Same dependency injection pattern
    logger      logger.Logger      // Same logging approach
}
```

2. **Maintain Consistent Error Handling**:
```go
// If existing code uses this pattern:
if err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// Don't introduce different patterns:
// WRONG
if err != nil {
    log.Error(err)  // Different error handling
    return nil, err
}
```

3. **Follow Existing Structure**:
```go
// If existing services follow this structure:
// 1. Validate input
// 2. Check business rules
// 3. Execute operation
// 4. Send notifications
// 5. Return result

// Your new service MUST follow the same flow
func (s *productService) CreateProduct(ctx context.Context, req CreateProductRequest) (*Product, error) {
    // 1. Validate input (following existing validation patterns)
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. Check business rules (using same patterns as existing services)
    if err := s.checkProductConstraints(ctx, req); err != nil {
        return nil, err
    }
    
    // 3. Execute operation (same transaction patterns)
    product, err := s.productRepo.Create(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }
    
    // 4. Send notifications (same event patterns)
    s.eventBus.Publish(ProductCreatedEvent{ProductID: product.ID})
    
    // 5. Return result
    return product, nil
}
```

### Unit Testing
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &mocks.UserRepository{}
    mockEmail := &mocks.EmailService{}
    
    svc, err := NewUserService(mockRepo, mockEmail)
    require.NoError(t, err)
    
    ctx := context.Background()
    req := CreateUserRequest{
        Email:    "test@example.com",
        Username: "testuser",
    }
    
    mockRepo.On("Create", ctx, mock.Anything).Return(nil)
    mockEmail.On("SendWelcome", ctx, "test@example.com").Return(nil)
    
    // Act
    user, err := svc.CreateUser(ctx, req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email.String())
    
    mockRepo.AssertExpectations(t)
    mockEmail.AssertExpectations(t)
}
```

### Table-Driven Tests
```go
func TestEmail_Validate(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty string", "", true},
        {"missing domain", "user@", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Testing
```go
// Place in test/integration/
func TestUserRepository_PostgreSQL(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := postgres.NewUserRepository(db)
    
    t.Run("Create and Find User", func(t *testing.T) {
        // Test actual database operations
    })
}
```

## API Design

### RESTful Endpoints
```go
// Follow RESTful conventions
func (c *UserController) RegisterRoutes(router *mux.Router) {
    router.HandleFunc("/users", c.ListUsers).Methods("GET")
    router.HandleFunc("/users/{id}", c.GetUser).Methods("GET")
    router.HandleFunc("/users", c.CreateUser).Methods("POST")
    router.HandleFunc("/users/{id}", c.UpdateUser).Methods("PUT")
    router.HandleFunc("/users/{id}", c.DeleteUser).Methods("DELETE")
}
```

### Request/Response DTOs
```go
// Input DTOs with validation tags
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=8"`
}

// Output DTOs hiding internal details
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}

// Mapper functions
func ToUserResponse(user *entities.User) UserResponse {
    return UserResponse{
        ID:        user.ID,
        Email:     user.Email.String(),
        Username:  user.Username,
        CreatedAt: user.CreatedAt,
    }
}
```

## Concurrency Patterns

### Goroutines and Channels
```go
// Use context for cancellation
func (s *userService) ProcessBatch(ctx context.Context, userIDs []string) error {
    errChan := make(chan error, len(userIDs))
    var wg sync.WaitGroup
    
    for _, id := range userIDs {
        wg.Add(1)
        go func(userID string) {
            defer wg.Done()
            
            select {
            case <-ctx.Done():
                errChan <- ctx.Err()
                return
            default:
                if err := s.processUser(ctx, userID); err != nil {
                    errChan <- fmt.Errorf("failed to process user %s: %w", userID, err)
                }
            }
        }(id)
    }
    
    wg.Wait()
    close(errChan)
    
    // Collect errors
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("batch processing failed: %v", errs)
    }
    
    return nil
}
```

### Worker Pools
```go
func (s *userService) ProcessWithWorkerPool(ctx context.Context, tasks []Task) {
    numWorkers := runtime.NumCPU()
    taskChan := make(chan Task, len(tasks))
    
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range taskChan {
                s.processTask(ctx, task)
            }
        }()
    }
    
    // Send tasks
    for _, task := range tasks {
        taskChan <- task
    }
    close(taskChan)
    
    wg.Wait()
}
```

## Database Patterns

### Repository Pattern
```go
// Interface in domain layer
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*entities.User, error)
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
    Create(ctx context.Context, user *entities.User) error
    Update(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id string) error
}

// Implementation in integration layer
type postgresUserRepository struct {
    db *sql.DB
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
    query := `SELECT id, email, username, created_at, updated_at 
              FROM users WHERE id = $1`
    
    var user entities.User
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID,
        &user.Email,
        &user.Username,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    return &user, nil
}
```

### Transaction Management
```go
func (s *userService) TransferCredits(ctx context.Context, fromID, toID string, amount int) error {
    return s.db.Transaction(func(tx *sql.Tx) error {
        // All operations within transaction
        if err := s.deductCredits(ctx, tx, fromID, amount); err != nil {
            return err
        }
        
        if err := s.addCredits(ctx, tx, toID, amount); err != nil {
            return err
        }
        
        return nil
    })
}
```

## Configuration Management

### Environment Configuration
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Host         string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
    Port         int           `env:"SERVER_PORT" envDefault:"8080"`
    ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
    WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &cfg, nil
}
```

## Observability

### Structured Logging
```go
// Use structured logging with context
func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
    logger := log.FromContext(ctx).With(
        "operation", "CreateUser",
        "email", req.Email,
    )
    
    logger.Info("creating user")
    
    user, err := s.userRepo.Create(ctx, req)
    if err != nil {
        logger.Error("failed to create user", "error", err)
        return nil, err
    }
    
    logger.Info("user created successfully", "user_id", user.ID)
    return user, nil
}
```

### Metrics
```go
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint", "status"},
    )
)

func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        wrapped := &responseWriter{ResponseWriter: w}
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start).Seconds()
        requestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.statusCode),
        ).Observe(duration)
    })
}
```

## Security Best Practices

### Input Validation
```go
func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // Validate input
    if err := validator.Validate(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Sanitize input
    req.Username = sanitize.HTML(req.Username)
    
    // Additional business validation
    if s.isReservedUsername(req.Username) {
        return errors.New("username is reserved")
    }
    
    return nil
}
```

### Authentication & Authorization
```go
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            
            claims, err := jwtService.ValidateToken(token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), "user_claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Performance Guidelines

### Optimization Principles
- Profile before optimizing
- Use benchmarks to measure improvements
- Cache frequently accessed data
- Use connection pooling for databases
- Implement pagination for large datasets
- Use appropriate data structures
- Minimize allocations in hot paths

### Benchmarking
```go
func BenchmarkUserService_GetUser(b *testing.B) {
    svc := setupService()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = svc.GetUser(ctx, "test-id")
    }
}
```

## Documentation Standards

### Package Documentation
```go
// Package user implements user management functionality.
// It provides services for user creation, authentication, and profile management.
//
// The package follows clean architecture principles with clear separation
// between business logic and infrastructure concerns.
package user
```

### API Documentation
```go
// CreateUser creates a new user account.
//
// It validates the input, checks for duplicate emails,
// hashes the password, and stores the user in the database.
// A welcome email is sent upon successful creation.
//
// Example:
//
//	req := CreateUserRequest{
//	    Email:    "user@example.com",
//	    Username: "johndoe",
//	    Password: "SecurePass123!",
//	}
//	user, err := service.CreateUser(ctx, req)
//
// Returns:
//   - *entities.User: The created user entity
//   - error: ErrEmailExists if email is taken, validation errors, or database errors
func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
    // Implementation
}
```

## Makefile Commands

```makefile
.PHONY: help build test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/app cmd/app/main.go

test: ## Run tests
	go test -v -race -cover ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./test/integration/...

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	goimports -w .

migrate-up: ## Run database migrations
	migrate -path scripts/migrations -database "${DATABASE_URL}" up

docker-build: ## Build Docker image
	docker build -f build/docker/Dockerfile -t app:latest .

run: ## Run the application
	go run cmd/app/main.go
```

## Development Workflow

### Mandatory BDD-First Approach

**CRITICAL**: Every new development MUST begin with creating a `.feature` file in `/test/integration/features/` before writing any implementation code.

1. **Feature Definition First**: Create a `.feature` file in `/test/integration/features/` describing the behavior using Gherkin syntax
2. **Use Existing Step Definitions**: Map scenarios to existing generic step definitions - DO NOT create new step definitions unless absolutely necessary
3. **Review Existing Code**: Study existing implementations for patterns, styles, and conventions before writing new code
4. **Domain Modeling**: Model domain entities and business rules following existing patterns
5. **Use Cases**: Define application use cases matching existing architectural patterns
6. **Infrastructure**: Implement infrastructure adapters consistent with existing code
7. **Testing**: Ensure all BDD scenarios pass, then add unit tests following existing test patterns
8. **Documentation**: Document following the existing documentation style
9. **Review**: Ensure code follows existing patterns and clean architecture principles

## Common Pitfalls to Avoid

- **Skipping BDD feature files** - NEVER write code without a .feature file first
- **Creating new BDD step definitions** - ALWAYS use existing generic steps
- **Ignoring existing code patterns** - ALWAYS study and follow existing implementations
- **Introducing new styles** - MAINTAIN consistency with existing code style
- **Reinventing existing functionality** - SEARCH for and reuse existing code first
- **Leaking domain logic** into controllers or repositories
- **Circular dependencies** between packages
- **Ignoring errors** or using blank identifier inappropriately
- **Over-engineering** simple CRUD operations
- **Tight coupling** to specific frameworks or libraries
- **Missing context** in long-running operations
- **Improper goroutine** lifecycle management
- **SQL injection** through string concatenation
- **Hardcoding** configuration values
- **Missing validation** at boundaries

## Pre-Development Checklist

Before writing ANY code, ensure:

- [ ] Feature file created in `/test/integration/features/`
- [ ] All scenarios use ONLY existing step definitions
- [ ] Reviewed similar existing features for patterns
- [ ] Identified reusable existing code components
- [ ] Confirmed following existing code style
- [ ] Mapped BDD scenarios to implementation plan
- [ ] No new step definitions will be created
- [ ] Error handling follows existing patterns
- [ ] Naming conventions match existing code
- [ ] Structure follows existing architectural patterns
