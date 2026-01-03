# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a RESTful API web service for a book/library management system written in Go. It's a backend service that provides book catalog management, user authentication with JWT, search functionality, and file export capabilities.

## Architecture

### Framework Stack
- **Gin** - HTTP web framework
- **XORM** - ORM for database operations (supports MySQL/SQLite)
- **Zap** - Structured logging
- **CLI** - `urfave/cli/v2` for command-line interface

### Directory Structure
- `book.go` - Main entry point using CLI framework
- `internal/` - Private implementation packages:
  - `cmd/` - CLI command implementations (api, export, extra)
  - `conf/` - Configuration management
  - `db/` - Database layer with XORM ORM
  - `initial/` - Application initialization
  - `logger/` - Structured logging with Zap
  - `middleware/` - JWT authentication middleware
  - `models/` - Business logic and data models
  - `modules/` - Additional modules (translation/i18n)
  - `route/` - HTTP route handlers and Gin framework setup
  - `util/` - Utility functions
- `conf/` - Configuration files (app.ini)
- `book.sql` - Database schema

### Key Patterns
- **Layered Architecture**: Route → Model → Database → Configuration
- **Dependency Injection**: Configuration loaded from INI files via `conf` package
- **Repository Pattern**: Database operations encapsulated in `db/` package
- **Middleware Pattern**: JWT authentication as Gin middleware

## Development Commands

### Building and Running
```bash
# Development mode (hot reload with air)
air

# Run API server directly
go run book.go api

# With custom config file
go run book.go -c /path/to/config.ini api

# Production build
go build -o book-api
./book-api api
```

### Configuration
- Primary config: `conf/app.ini`
- Supports custom config path via CLI flag `-c`
- Environment variable `BOOK_WORK_DIR` sets working directory for config resolution
- If `BOOK_WORK_DIR` is not set, config must be in same directory as binary

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test ./internal/route -run TestBookHandler

# Run tests with verbose output
go test -v ./...
```

### Dependency Management
```bash
# Add new dependency
go get -u example.com/component

# Tidy dependencies
go mod tidy

# Check for updates
go list -m -u all
```

### Performance Profiling
```bash
# CPU profiling (20 seconds)
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/profile

# Memory profiling
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/heap

# Goroutine profiling
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/goroutine

# Web UI for profiling results (requires graphviz)
go tool pprof -http=:9966 log/cpu.pprof
```

## Database

### Supported Databases
- MySQL (primary)
- SQLite

### Schema Management
- Initial schema: `book.sql`
- XORM handles migrations automatically
- Connection pooling configured in `conf/app.ini`

### Key Tables
- `book` - Main book information
- `category` - Book categories
- `chapter` - Book chapters
- `content` - Chapter content
- `user` - User accounts
- `volume` - Book volumes

## Configuration

### Environment Variables
- `BOOK_WORK_DIR` - Working directory for config file resolution
- Database credentials and other settings in `conf/app.ini`

### INI File Structure
```ini
[server]
HTTP_PORT = 6767
RUN_MODE = debug

[database]
DB_TYPE = mysql
HOST = 127.0.0.1:3306
NAME = book
USER = root
PASSWD =
SSL_MODE = disable

[i18n]
LANGS = en-US,zh-CN
NAMES = English,简体中文
```

## API Features

### Authentication
- JWT-based authentication via `gin-jwt/v2`
- Middleware in `internal/middleware/auth.go`
- Token refresh support

### Core Functionality
- Book CRUD operations
- Author management
- Category management
- Chapter and volume management
- Search functionality
- Export to various formats (PDF, EPUB, etc.)

### Additional Features
- CORS support
- GZIP compression
- Structured logging with file rotation
- Internationalization (i18n) support
- Performance profiling endpoints

## Development Notes

### Code Style
- Mixed Chinese and English comments
- Uses `internal/` package for private implementation
- Error handling with `github.com/pkg/errors`
- Structured logging with Zap

### Hot Reload
- Configured with Air (`.air.toml`)
- Excludes test files and vendor directory
- Builds to `./tmp/main`

### Testing
- Test files use `_test.go` suffix
- Located alongside source files
- Use `github.com/stretchr/testify` for assertions

### CI/CD
- GitHub Actions workflow in `.github/workflows/go.yml`
- Builds with Go 1.20
- Artifacts named `book-api`

## Common Issues

### Configuration File Not Found
When running `go run book.go`, the `conf/app.ini` file may not be found because Go doesn't copy the conf directory during compilation. Solutions:
1. Set `BOOK_WORK_DIR` environment variable: `BOOK_WORK_DIR=/path/to/work go run book.go`
2. Use `-c` flag to specify config path: `go run book.go -c /path/to/config.ini api`
3. Add `export BOOK_WORK_DIR=/path/to/work` to your shell profile

### Stopping the Server
- Use `CTRL+C`
- Or kill process: `ps -ef | grep book | grep -v grep | awk '{print $2}' | xargs kill -15`

## Related Projects
- Frontend: [book-frontend](https://github.com/zgia/book-frontend.git)
- Inspired by: [gogs](https://github.com/gogs/gogs) architecture