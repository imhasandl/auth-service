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

Creates a new user account with the provided credentials. The service validates the input data, securely hashes the password using bcrypt, generates a verification code, and sends it to the user's email address for account verification. The user information is stored in the database with verification status set to false until the user completes email verification.

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

Authenticates a user using their email/username and password. The service validates credentials against the stored hashed password, generates a JWT access token (valid for 1 hour) and a refresh token (valid for 7 days) upon successful authentication. The tokens are used for subsequent authorized API calls and maintaining user sessions.

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

---

### VerifyEmail

Verifies a user's email address using the verification code sent to their email.

#### Request format

```json
{
  "email": "user email",
  "verification_code": 1234 // the numeric code sent to the user's email
}
```

#### Response format

```json
{
  "success": true, // boolean indicating if verification was successful
  "message": "Email verified successfully" // status message
}
```

---

### SendVerifyCodeAgain

Requests a new verification code when the original code expires or gets lost.

#### Request format

```json
{
  "email": "user email"
}
```

#### Response format

```json
{
  "success": true, // boolean indicating if new code was sent successfully
  "message": "new verification code sent" // status message
}
```

---

### RefreshToken

Generates a new access token using a valid refresh token.

#### Request format

```json
{
  "refresh_token": "user's refresh token"
}
```

#### Response format

```json
{
  "access_token": "new JWT access token",
  "refresh_token": "new refresh token",
  "expiry_time": "timestamp when the access token will expire",
  "error": "" // contains error message if any
}
```

---

### Logout

Invalidates a user's refresh token to log them out.

#### Request format

```json
{
  "refresh_token": "user's refresh token"
}
```

#### Response format

```json
{
  "success": true, // boolean indicating if logout was successful
  "message": "User logged out complete" // status message
}
```

----

## Running the Service

```bash
go run cmd/main.go
```

## Docker Support

The service can be run as part of a Docker Compose setup along with other microservices. When using Docker, make sure to use the Docker Compose specific DB_URL configuration.

