# Setup Guide

## Prerequisites

1. **Install Go**
   - Download Go from [https://golang.org/dl/](https://golang.org/dl/)
   - Add Go to your system PATH
   - Verify installation: `go version`

2. **Install PostgreSQL**
   - Download PostgreSQL from [https://www.postgresql.org/download/](https://www.postgresql.org/download/)
   - Create a new database named `topup_game_db`

3. **Environment Setup**
   - Copy `.env.example` to `.env`
   - Update the values in `.env` with your configuration

## Installation Steps

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up the database**
   ```sql
   CREATE DATABASE topup_game_db;
   ```

4. **Configure environment variables**
   Create a `.env` file with the following content:
   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_USER=your_postgres_user
   DB_PASSWORD=your_postgres_password
   DB_NAME=topup_game_db
   DB_PORT=5432

   # JWT Configuration
   JWT_SECRET=your-secret-key-here

   # VIP Reseller API Configuration
   VIP_RESELLER_API_KEY=your-api-key
   VIP_RESELLER_USER_ID=your-user-id
   VIP_RESELLER_BASE_URL=https://vip-reseller.co.id/api
   ```

5. **Run the application**
   ```bash
   go run cmd/main.go
   ```

## Testing

1. **Register a new admin user**
   ```bash
   curl -X POST -H "Content-Type: application/json" \
        -d '{"email":"admin@example.com","password":"admin123","role":"admin"}' \
        http://localhost:8080/api/auth/register
   ```

2. **Login**
   ```bash
   curl -X POST -H "Content-Type: application/json" \
        -d '{"email":"admin@example.com","password":"admin123"}' \
        http://localhost:8080/api/auth/login
   ```

3. **Access the application**
   - Main website: http://localhost:8080
   - Admin dashboard: http://localhost:8080/admin
   - Transaction status: http://localhost:8080/transaction/:invoice

## Project Structure

```
├── cmd/
│   └── main.go           # Application entry point
├── config/
│   └── config.go         # Configuration management
├── internal/
│   ├── handler/          # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── model/           # Database models
│   ├── repository/      # Database operations
│   ├── service/         # Business logic
│   └── router/          # Route definitions
├── templates/           # HTML templates
├── .env                # Environment variables
└── README.md          # Project documentation
```

## Common Issues

1. **Database Connection Issues**
   - Ensure PostgreSQL is running
   - Verify database credentials in `.env`
   - Check if database exists

2. **Go Module Issues**
   - Run `go mod tidy` to clean up dependencies
   - Ensure Go version 1.21 or later is installed

3. **Permission Issues**
   - Ensure proper file permissions for `.env`
   - Check database user permissions

## Support

For any issues or questions, please:
1. Check the common issues section above
2. Review the error logs
3. Contact the development team
