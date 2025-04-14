[![CI](https://github.com/imhasandl/auth-service/actions/workflows/ci.yml/badge.svg)](https://github.com/imhasandl/auth-service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/imhasandl/auth-service)](https://goreportcard.com/report/github.com/imhasandl/auth-service)
[![GoDoc](https://godoc.org/github.com/imhasandl/auth-service?status.svg)](https://godoc.org/github.com/imhasandl/auth-service)
[![Coverage](https://codecov.io/gh/imhasandl/auth-service/branch/main/graph/badge.svg)](https://codecov.io/gh/imhasandl/auth-service)
[![Go Version](https://img.shields.io/github/go-mod/go-version/imhasandl/auth-service)](https://golang.org/doc/devel/release.html)

---

## Authentification Service

#### Core Functionality

User Registration and Login: The service provides endpoints for users to create new accounts and log in with their credentials.

Token-Based Authentication This allows other services to securely identify users without needing to directly interact with the authentication database.

Password Hashing: Robust password hashing algorithms (bcrypt) being used to securely store user passwords in the database, preventing plain-text storage.

Database Integration: The service will undoubtedly interact with a database to store user credentials and related information using Postgresql database.

---

## Prerequisites

- Go 1.20 or later
- PostgreSQL database

## Configuration

Create a `.env` file in the root directory with the following variables:

```env
PORT=":YOUR_GRPC_PORT"
DB_URL="postgres://username:password@host:port/database?sslmode=disable"
# DB_URL="postgres://username:password@db:port/database?sslmode=disable" // FOR DOCKER COMPOSE
EMAIL="Company email for sending email notification for account validation"
EMAIL_SECRET="email pass phrase"
```

This service uses Goose for database migrations:

```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir migrations postgres "YOUR_DB_CONNECTION_STRING" up
```

## gRPC Methods

The service implements the following gRPC methods:

---

### Register

Registers users credentials and stores them into a database.

#### Request format

```json
{
  "email": "user email",
  "password": "user's password (it will we encrypted and sent to database)",
  "username": "user username"
}
```

#### Response format

```json
{
  "user": {
    "id": "UUID of a user",
    "created_at": "time when user signed in",
    "updated_at": "when user changes something(changes username, password etc.)",
    "email": "user email",
    "username": "user username",
    "subscribers": "this is an array of uuid of users that subscribed to this user",
    "subscribed_to": "this is an array aswell and stores the id's of currently subscribed users by the current user",
    "is_premium": "TRUE or FALSE, there might be some subscription system and we can use this field",
    "verification_code": "if user want to verify its account, the 6 digit code will be sent to user email",
    "is_verified": "bool value that defines if user verified TRUE it's account on not FALSE"
  }
}
```

> **Note:** As you can see there is no password field, because we don't response user information for security reasons.

---

### Login

Let the user to log in.

#### Request format

```json
{
  "identifier": "username or email",
  "password": "user password"
}
```

#### Response format

```json
{
  "user": {
    "id": "UUID of a user",
    "created_at": "time when user signed in",
    "updated_at": "when user changes something(changes username, password etc.)",
    "email": "user email",
    "username": "user username",
    "subscribers": "this is an array of uuid of users that subscribed to this user",
    "subscribed_to": "this is an array aswell and stores the id's of currently subscribed users by the current user",
    "is_premium": "TRUE or FALSE, there might be some subscription system and we can use this field",
    "verification_code": "if user want to verify its account, the 6 digit code will be sent to user email",
    "is_verified": "bool value that defines if user verified TRUE it's account on not FALSE"
  },
  "token": "user token for authorization",
  "refresh_token": "token"
}
```



