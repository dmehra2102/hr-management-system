# HR Management System (Work In-Progress)

A complete, production-grade Human Resources Management System built with Go, gRPC, and Protocol Buffers.

## 🚀 Features

### Core HR Modules
- **Employee Management**: Complete CRUD operations for employee records
- **Department Management**: Organization structure and department hierarchies  
- **Leave Management**: Leave requests, approvals, and balance tracking
- **Performance Management**: Performance reviews, goal setting, and competency tracking
- **Authentication**: JWT-based authentication with role-based permissions

### Technical Features
- **gRPC API**: High-performance API using Protocol Buffers
- **PostgreSQL Database**: Reliable data storage with migrations
- **JWT Authentication**: Secure token-based authentication
- **Role-Based Access Control**: Admin, HR, Manager, and Employee roles
- **Structured Logging**: JSON-formatted logs with request tracing
- **Database Migrations**: Version-controlled schema changes
- **Docker Support**: Containerized deployment
- **Middleware**: Authentication, logging, and recovery middleware
- **Production-Ready**: Configuration management, health checks, graceful shutdown

## 🏗️ Architecture

```
├── api/proto/v1/          # Protocol Buffer definitions
├── cmd/server/            # Application entry point
├── internal/              # Private application code
│   ├── auth/              # Authentication service
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and migrations
│   ├── employee/          # Employee service (models, repository, service, handler)
│   ├── department/        # Department service
│   ├── leave/             # Leave management service
│   ├── performance/       # Performance management service
│   └── middleware/        # gRPC middleware
├── pkg/                   # Shared/reusable packages
│   ├── logger/            # Structured logging
│   ├── response/          # API response utilities
│   └── validator/         # Input validation
└── scripts/               # Build and deployment scripts
```

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **API**: gRPC with Protocol Buffers
- **Database**: PostgreSQL 15+
- **Authentication**: JWT tokens
- **Logging**: Logrus with structured JSON logging
- **Configuration**: Viper with environment variables
- **Database ORM**: GORM v2
- **Migrations**: golang-migrate
- **Containerization**: Docker & Docker Compose
- **Password Hashing**: bcrypt

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+ 
- Docker and Docker Compose
- Protocol Buffers compiler (`protoc`)
- Make (optional, for using Makefile)

## 🚀 Quick Start

### 1. Clone the repository
```bash
git clone <repository-url>
cd hr-management-system
```

### 2. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start with Docker Compose (Recommended)
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f hr-api
```

### 4. Manual Setup (Development)

#### Install dependencies
```bash
go mod download
make install-tools
```

#### Start PostgreSQL
```bash
# Using Docker
docker run -d \
  --name hr-postgres \
  -e POSTGRES_USER=hruser \
  -e POSTGRES_PASSWORD=hrpassword \
  -e POSTGRES_DB=hrmanagement \
  -p 5432:5432 \
  postgres:15-alpine
```

#### Generate Proto files
```bash
make proto
```

#### Run database migrations
```bash
make migrate-up
```

#### Start the server
```bash
make run
# or
go run cmd/server/main.go
```

## 🔧 Configuration

Configuration is managed through environment variables. See `.env.example` for all available options.

### Key Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | hruser | Database user |
| `DB_PASSWORD` | hrpassword | Database password |
| `DB_NAME` | hrmanagement | Database name |
| `JWT_SECRET` | - | JWT signing secret (required) |
| `GRPC_PORT` | 9090 | gRPC server port |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `APP_ENV` | development | Environment (development, staging, production) |

## 🧪 Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
go test -cover ./...
```

## 📖 API Documentation

The system provides gRPC services for:

### Employee Service
- `CreateEmployee` - Create new employee
- `GetEmployee` - Get employee by ID
- `UpdateEmployee` - Update employee information
- `DeleteEmployee` - Soft delete employee
- `ListEmployees` - List employees with pagination and filtering
- `GetEmployeesByDepartment` - Get employees by department

### Department Service
- `CreateDepartment` - Create new department
- `GetDepartment` - Get department by ID
- `UpdateDepartment` - Update department information
- `DeleteDepartment` - Delete department
- `ListDepartments` - List departments

### Authentication Service
- `Login` - Authenticate user
- `RefreshToken` - Refresh access token
- `Logout` - Logout user
- `ValidateToken` - Validate access token
- `ChangePassword` - Change user password

### Leave Service
- `CreateLeaveRequest` - Create leave request
- `GetLeaveRequest` - Get leave request by ID
- `UpdateLeaveRequest` - Update leave request
- `ListLeaveRequests` - List leave requests
- `ApproveLeaveRequest` - Approve leave request
- `RejectLeaveRequest` - Reject leave request
- `GetEmployeeLeaveBalance` - Get employee leave balance

### Performance Service
- `CreatePerformanceReview` - Create performance review
- `GetPerformanceReview` - Get performance review by ID
- `UpdatePerformanceReview` - Update performance review
- `ListPerformanceReviews` - List performance reviews
- `SubmitPerformanceReview` - Submit performance review

## 🔒 Authentication & Authorization

The system uses JWT-based authentication with role-based access control:

### Roles
- **ADMIN**: Full system access
- **HR**: HR operations and employee management
- **MANAGER**: Team management and performance reviews
- **EMPLOYEE**: Self-service operations

### Authentication Flow
1. User logs in with email/password
2. System returns JWT access token and refresh token
3. Client includes JWT in `Authorization: Bearer <token>` header
4. Server validates JWT and extracts user permissions
5. Access is granted based on role and permissions

## 🚢 Deployment

### Docker Deployment
```bash
# Build image
docker build -t hr-management-system .

# Run with environment variables
docker run -d \
  --name hr-api \
  -p 9090:9090 \
  -e DB_HOST=your-db-host \
  -e JWT_SECRET=your-secret \
  hr-management-system
```

### Kubernetes Deployment
Example Kubernetes manifests are provided in the `deployments/` directory.

## 🔍 Monitoring & Observability

### Logging
- Structured JSON logging with Logrus
- Request ID tracking for distributed tracing
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Component-specific logging (service, repository, handler)

### Health Checks
- Database connectivity health check
- Graceful shutdown handling
- Connection pool monitoring

## 🧹 Development

### Code Structure
- Clean Architecture principles
- Domain-driven design
- Repository pattern for data access
- Service layer for business logic
- Handler layer for gRPC endpoints

### Adding New Features
1. Define Protocol Buffers in `api/proto/v1/`
2. Generate Go code: `make proto`
3. Implement models in `internal/{service}/models.go`
4. Create repository interface and implementation
5. Implement service layer with business logic
6. Create gRPC handler
7. Add to service registration in `main.go`
8. Add tests

### Database Migrations
```bash
# Create new migration
migrate create -ext sql -dir internal/database/migrations -seq your_migration_name

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```


