# Top Up Game Website

A Go-based game top-up website with clean architecture.

## Tech Stack

- Golang 1.21+
- Gin Web Framework
- GORM with PostgreSQL
- JWT Authentication
- TailwindCSS
- HTML Templates

## Project Structure

```
├── cmd
│   └── main.go
├── config
│   └── config.go
├── internal
│   ├── handler
│   ├── service
│   ├── repository
│   ├── model
│   └── router.go
├── templates
├── static
├── .env
└── README.md
```

## Setup Instructions

1. Clone the repository
2. Copy `.env.example` to `.env` and update the values
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Features

- User authentication (Admin/Reseller)
- Product management
- Transaction processing
- Integration with VIP Reseller API
- Real-time transaction status checking

## API Endpoints

### Public Endpoints
- `GET /products` - List all products
- `POST /checkout` - Process checkout
- `GET /transaction/:invoice` - Check transaction status

### Protected Endpoints (Admin/Reseller)
- `POST /admin/login` - Admin login
- `GET /admin/products` - View products
- `POST /admin/products` - Add product
- `GET /admin/transactions` - View all transactions
- `GET /admin/transactions/:id` - View transaction details

## Database Models

### User
- ID
- Email
- Password
- Role (guest, admin, reseller)
- Timestamps

### Product
- ID
- Name
- Category
- Price
- Timestamps

### Transaction
- ID
- UserID (nullable)
- ProductID
- Method
- Invoice (unique)
- Status
- Timestamps
