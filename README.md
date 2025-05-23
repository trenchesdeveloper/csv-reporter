# CSV Reporter

A modern Go web application for generating and managing CSV reports with robust user authentication.

## Project Overview

CSV Reporter is a RESTful API service built with Go that enables users to generate, view, and manage CSV reports. The application features a secure authentication system with JWT tokens and refresh token rotation.

## Features

- **User Authentication**
  - Secure signup and login with password hashing
  - JWT-based authentication
  - Refresh token rotation for enhanced security

- **API Documentation**
  - Swagger UI for interactive API documentation
  - Well-documented endpoints

- **Database**
  - PostgreSQL for data persistence
  - Database migrations for version control
  - SQL query generation with sqlc

- **Modern Go Practices**
  - Clean architecture
  - Dependency injection
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
