# DigiCert Book API

A RESTful API for managing books, users, and borrowing system. Built with Go, PostgreSQL, Docker, and AWS CloudWatch metrics.

---

## Features

- **User Registration & Login** (JWT-based authentication)
- **Admin Registration**
- **Book CRUD** (Admin only)
- **User Profile Management**
- **Borrow & Return Books**
- **Admin User Management**
- **CloudWatch Metrics Integration**
- **Swagger/OpenAPI Documentation**
- **Health & Readiness Endpoints**
- **Graceful Shutdown**

---
- **Architecture**
[User Client] --> [Ingress / API Gateway] --> [library-api (HTTP server)]
                                                  |
                                                  v
                                       +----------------------+
                                       |  HTTP Handlers       |
                                       |  (books, bookings,   |
                                       |   users, auth)       |
                                       +----------------------+
                                                  |
                                                  v
                                       +----------------------+
                                       |  Service Layer       |
                                       |  (business logic,    |
                                       |   validation, rules) |
                                       +----------------------+
                                                  |
                                                  v
                                       +----------------------+
                                       |  Repo Layer (pgx)    |
                                       |  (books, users,      |
                                       |   bookings repos)    |
                                       +----------------------+
                                                  |
                                                  v
                                          [Postgres DB]
                                          (migrations run at deploy)
## Deployment

### EC2 Instance

- **Public IP:** `98.92.152.54`
- **Private IP:** `172.31.70.59`
-- **DNS:** http://library-api-alb-916820835.us-east-1.elb.amazonaws.com
- The API is deployed and accessible via the public IP (ensure security group allows inbound traffic on the API port, e.g. `8080`).

---

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL

### Clone the Repository

```bash
git clone https://github.com/<your-username>/digicert.git
cd digicert
```

### Environment Variables

Create a `.env` file in the project root:

```properties
POSTGRES_USER=postgres
POSTGRES_PASSWORD=example
POSTGRES_DB=digicert
DATABASE_URL=postgres://postgres:example@db:5432/digicert?sslmode=disable
PORT=8080
ENABLE_CLOUDWATCH=false
AWS_REGION=us-east-1
CW_LOG_GROUP=/aws/ec2/library-api
CW_LOG_STREAM=library-api-dev
WT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_TTL=24h
```

---

## Running Locally

### With Docker Compose

```bash
docker-compose up
```

- API available at `http://localhost:8080`
- PostgreSQL runs in a container

### Without Docker

1. Start PostgreSQL locally.
2. Update `DATABASE_URL` in `.env` to use `localhost`.
3. Run:

    ```bash
    go run ./cmd/library-api/main.go
    ```

---

## API Documentation

Swagger UI available at `/swagger/index.html` (if enabled).

---

## Endpoints

### Health

- `GET /healthz` — Health check
- `GET /readyz` — Readiness check

### Auth

- `POST /auth/register` — Register user
- `POST /auth/admin-register` — Register admin
- `POST /auth/login` — Login
- `POST /auth/refresh` — Refresh JWT

### Users

- `GET /users/me` — Get profile
- `PUT /users/me` — Update profile

### Books

- `GET /books` — List books
- `GET /books/{id}` — Get book details

### Admin (Protected)

- `POST /admin/books` — Create book
- `PUT /admin/books/{id}` — Update book
- `DELETE /admin/books/{id}` — Delete book
- `GET /admin/users` — List users
- `GET /admin/users/{id}` — Get user
- `DELETE /admin/users/{id}` — Delete user
- `GET /admin/bookings` — List all bookings

### Borrowing

- `GET /bookings` — List my bookings
- `POST /bookings` — Borrow book
- `GET /bookings/{id}` — Get booking
- `POST /bookings/{id}/return` — Return book

---

## CloudWatch Metrics

- Metrics are sent to AWS CloudWatch if `ENABLE_CLOUDWATCH=true`.
- For local development, set `ENABLE_CLOUDWATCH=false` to disable metrics/logs to AWS.

---

## Graceful Shutdown

The server supports graceful shutdown on `Ctrl+C`.

---

## Architecture Decisions

See the `/adr/` directory for Architecture Decision Records (ADRs) documenting key technical choices (Go, PostgreSQL, JWT, Docker, CloudWatch, etc.).

---

## License

Apache 2.0

---

## Contributing

Pull requests welcome! Please open issues for bugs or feature requests.

---

## Contact

API Support: [support@swagger.io](mailto:support@swagger.io)
