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
