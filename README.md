# CSV Reporter

A modern Go web application for generating and managing CSV reports with robust user authentication.

## Project Overview

CSV Reporter is a RESTful API service built with Go that enables users to generate, view, and manage CSV reports. The application features a secure authentication system with JWT tokens and refresh token rotation.

## Features

- **User Authentication**
  - Secure signup and login with password hashing (bcrypt)
  - JWT-based authentication with access and refresh tokens
  - Refresh token rotation for enhanced security
  - Token blacklisting for logout functionality
  - SHA256 + bcrypt for secure token storage

- **Report Generation**
  - Support for multiple report types (monsters, weapons, armor)
  - Asynchronous report generation via queue system
  - Status tracking (requested, processing, completed, failed)
  - Secure download URLs with expiration
  - Error handling and reporting

- **Cloud Integration**
  - AWS S3 for report storage
  - SQS for message queuing
  - Presigned URLs for secure, time-limited file access
  - Infrastructure as Code with Terraform

- **API Documentation**
  - Swagger UI for interactive API documentation
  - Well-documented endpoints and response schemas

- **Database**
  - PostgreSQL for data persistence
  - Database migrations for version control
  - SQL query generation with sqlc
  - Dynamic queries using named parameters
  - Connection pooling for better performance

- **Modern Go Practices**
  - Clean architecture and separation of concerns
  - Dependency injection
  - Context-aware request handling
  - Middleware-based processing pipeline
  - Comprehensive logging and error tracking
  - Comprehensive test coverage

- **DevOps & Development Workflow**
  - Docker for containerization and local development
  - Makefile for streamlined commands
  - CI/CD potential with organized project structure
  - Environment-based configuration
  - Comprehensive test coverage

## Tech Stack

- **Backend**: Go 1.23
- **Web Framework**: Chi router
- **Database**: PostgreSQL
- **ORM/Query Builder**: sqlc
- **Authentication**: JWT
- **API Documentation**: Swagger
- **Logging**: Zap
- **Configuration**: Viper
- **Testing**: Go's testing package with Testify

## Project Structure

```
.
├── cmd
│   ├── api            # Main API server code
│   └── worker         # Worker service for processing reports
├── db
│   ├── migrations     # Database migration files
│   ├── queries        # SQL queries for sqlc
│   └── sqlc           # Generated Go code from SQL
├── helpers            # Utility functions
├── middleware         # HTTP middleware components
├── models             # Data models
├── reports            # Report generation logic
├── terraform          # Infrastructure as code
├── tests              # Test files
├── token              # Token authentication management
├── Makefile           # Project commands
└── docker-compose.yml # Local development setup
```

## Installation & Setup

### Prerequisites
- Go 1.19+
- PostgreSQL
- Docker and Docker Compose
- AWS account (for production deployment)
- Terraform (for infrastructure provisioning)

### Local Development Setup

1. Clone the repository
   ```
   git clone https://github.com/trenchesdeveloper/csv-reporter.git
   cd csv-reporter
   ```

2. Start the PostgreSQL database
   ```
   make makepostgres
   ```

3. Create database
   ```
   make createdb
   ```

4. Run migrations
   ```
   make migrateup
   ```

5. Generate SQL code
   ```
   make sqlc
   ```

6. Start the API server
   ```
   make server
   ```

7. In a separate terminal, start the worker
   ```
   make startWorker
   ```

### Environment Variables

Create a `.env` file in the project root with the following variables:

```
APP_ENV=dev
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/goflow?sslmode=disable
DB_TEST_SOURCE=postgresql://root:secret@localhost:5433/goflow_test?sslmode=disable
SERVER_ADDRESS=:8080
TOKEN_SYMMETRIC_KEY=your_symmetric_key_min_32_chars
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
S3_BUCKET=your-s3-bucket-name
SQS_QUEUE=your-sqs-queue-name
```

## API Usage

### Authentication

1. **User Registration**
   ```
   POST /api/v1/signup
   ```
   Request body:
   ```json
   {
     "email": "user@example.com",
     "password": "securepassword"
   }
   ```

2. **User Login**
   ```
   POST /api/v1/signin
   ```
   Request body:
   ```json
   {
     "email": "user@example.com",
     "password": "securepassword"
   }
   ```
   Response includes access and refresh tokens

3. **Token Refresh**
   ```
   POST /api/v1/refresh-token
   ```
   Request body:
   ```json
   {
     "refresh_token": "your-refresh-token"
   }
   ```

### Report Management

1. **Create Report**
   ```
   POST /api/v1/reports
   ```
   Request body:
   ```json
   {
     "report_type": "monsters"
   }
   ```

2. **Get Report**
   ```
   GET /api/v1/reports/:reportId
   ```

## Testing

Run tests with:
```
make test
```

## Deployment

### Infrastructure Provisioning with Terraform

1. Initialize Terraform
   ```
   make tf-init
   ```

2. Plan deployment
   ```
   make tf-plan
   ```

3. Apply infrastructure changes
   ```
   make tf-apply
   ```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
