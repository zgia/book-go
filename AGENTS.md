# AGENTS.md

This file provides guidance to AI when working with code in this repository.

## Project Overview

This is a RESTful API web service for a book/library management system written in Go. It's a backend service that provides book catalog management (books, authors, categories, volumes, chapters, contents), user authentication with JWT, search functionality, and text export. The companion frontend is [book-frontend](https://github.com/zgia/book-frontend.git).

## Architecture

### Framework Stack
- **Gin** - HTTP web framework
- **XORM** - ORM for database operations (supports MySQL/SQLite)
- **Zap** - Structured logging (rotated with lumberjack)
- **CLI** - `urfave/cli/v2` for command-line interface
- **gin-jwt/v2** - JWT authentication middleware
- **golang.org/x/text** - Internationalization via the in-repo `translation` module (en-US, zh-CN)

### Directory Structure
- `book.go` - Main entry point: CLI app, commands, global flags, and `--daemon` handling
- `conf/` - Embedded default config (`app.ini`) and locale files
- `book.sql` - MySQL initial schema for database setup
- `internal/` - Private implementation packages:
  - `cmd/` - CLI command implementations (api, export, extra, fts)
  - `conf/` - Configuration loading (embedded defaults + optional custom override)
  - `daemon/` - Self-daemonization used by `book api --daemon`
  - `db/` - Database layer with XORM ORM; `fts.go` holds the SQLite 2-gram full-text index
  - `initial/` - Application initialization
  - `logger/` - Structured logging with Zap
  - `middleware/` - JWT authentication middleware
  - `models/` - Business logic and data models
  - `modules/` - Additional modules (translation/i18n)
  - `route/` - HTTP route handlers and Gin framework setup
  - `util/` - Utility functions

### Key Patterns
- **Layered Architecture**: Route → Model → Database → Configuration
- **Dependency Injection**: Configuration loaded from INI files via `conf` package
- **Repository Pattern**: Database operations encapsulated in `db/` package
- **Middleware Pattern**: JWT authentication as Gin middleware
- **Self-daemonization**: `main` spawns a detached copy of itself when `--daemon` is given, then exits

## Development Commands

### Building and Running
```bash
# Development mode (hot reload with air; builds ./tmp/main and runs "api")
air

# Run API server directly (needs BOOK_WORK_DIR or -c, see Configuration)
BOOK_WORK_DIR=$(pwd) go run book.go api

# With custom config file
BOOK_WORK_DIR=$(pwd) go run book.go -c custom/conf/app.ini api

# Production build (FTS5 requires the build tag; mattn/go-sqlite3 does not ship it by default)
go build -tags sqlite_fts5 -o book .
./book api

# Run in background (detaches, no nohup needed)
./book api --daemon
```

### Lint and Tests
```bash
gofmt -l .                 # Format check
go vet ./...               # Static analysis
go test ./...              # All tests
go test -v ./internal/daemon/ -run TestRequested   # Single test
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
Requires `[server] ENABLE_PPROF = true`.

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

## CLI Commands

```
book [global options] command [command options]
```

Global flags:
- `-c, --config <path>` - Custom config file path

Commands:
- `api` - Run the HTTP API server; `--daemon` / `-d` runs it in background
- `export -b <book_id> [-p <dir>]` - Export books after the given book ID to text files
  (`-b` is required, `-p` defaults to `/tmp` and must be writable)
- `extra` - Print application info: version, app path, config path, dirs, log dir
- `fts rebuild [-b 500]` - Rebuild the SQLite 2-gram full-text index (`content_fts`)
- `fts stat` - Show index coverage (indexed chapters vs total contents)

## Configuration

### Loading Rules
1. The default `conf/app.ini` is embedded into the binary at build time (`conf/embed.go`).
2. An optional custom file is layered on top of it:
   - `-c/--config` path if given (also sets `BOOK_WORK_DIR` to the portion before `/custom/conf/app.ini`);
   - otherwise `<WORK DIR>/custom/conf/app.ini`.
3. **Work dir** = `BOOK_WORK_DIR` env var, else the directory of the running binary.

### Environment Variables
- `BOOK_WORK_DIR` - Working directory for config resolution, logs (`<work dir>/log`), and `custom/` dir
- `BOOK_DAEMONIZED` - Internal marker (`=1`) set on the detached child process by `internal/daemon`; never set it manually

### INI File Structure (real keys)

```ini
BRAND_NAME = zGia! Book Library
RUN_MODE = dev                 ; dev / prod / test; "prod" switches Gin to release mode

[server]
EXTERNAL_URL = %(PROTOCOL)s://%(DOMAIN)s:%(HTTP_PORT)s/
DOMAIN = localhost
PROTOCOL = http                ; http / https / fcgi / unix
HTTP_ADDR = 0.0.0.0
HTTP_PORT = 6767
ENABLE_GZIP = false
ENABLE_PPROF = false

[database]
TYPE = mysql                   ; mysql / sqlite / sqlite3
HOST = 127.0.0.1:3306          ; MySQL host:port, or SQLite db file path
NAME = library                 ; MySQL db name; SQLite fallback path when HOST is empty
USER = library
PASSWORD =
MAX_OPEN_CONNS = 30
MAX_IDLE_CONNS = 30

[http]
ACCESS_CONTROL_ALLOW_ORIGIN = *
JWT_SECRET_KEY = <change-me>

[api]
RESPONSE_ITEMS = 10
MAX_RESPONSE_ITEMS = 50

[log]
ROOT_PATH =                    ; default: <WORK DIR>/log
MODE = console,file
LEVEL = Debug
[log.file]
LOG_NAME = book.log
MAX_SIZE = 100
MAX_AGE = 7
MAX_BACKUPS = 10

[i18n]
LANGS = en-US,zh-CN
NAMES = English,简体中文
```

Note: `[log]` reads `ROOT_PATH` in code (`conf.LogDir()`), while the shipped `conf/app.ini` comments use `ROOT_DIR`; leave it empty to use `<WORK DIR>/log`.

## Database

### Supported Databases
- MySQL (primary; unix socket supported when `HOST` starts with `/`)
- SQLite / SQLite3 (`TYPE = sqlite`, `HOST` = db file path)

### Schema Management
- Initial schema: `book.sql` (MySQL DDL, apply manually; the code does not auto-sync tables)
- SQLite is supported by the code but has no bundled schema script; derive one from `book.sql` if needed
- XORM is used for queries only; `db.NewEngine` just connects and configures the pool
- Connection pooling configured in `[database]`; for SQLite the open count is clamped to 1..8 (default 4) and reused as the idle count
- SQLite connections are opened in WAL mode (`_journal_mode=wal&_busy_timeout=5000&_synchronous=normal&_foreign_keys=1&mode=rwc`); startup reads the pragmas back into the log
- Full-text search: `internal/db/fts.go` maintains `content_fts`, an FTS5 table over `content.txt` using space-separated 2-grams, `content=''` (contentless) and `detail=none`. Build with `-tags sqlite_fts5` or the module is missing; when the table is absent the search silently falls back to `LIKE '%kw%'`. `db.InitFTS()` probes it at startup and logs the state. Chapter insert/update keeps it in sync; a single-character, punctuation, or whitespace query also falls back to LIKE.

### Key Tables
- `author` - Authors (name, former_name)
- `book` - Main book info (title, alias, author_id, category_id, summary, source, latest, rate, word_count, is_finished, cover)
- `category` - Book categories (parent_id, title, is_hidden)
- `chapter` - Chapters (book_id, volume_id, title, word_count)
- `content` - Chapter content (chapter_id, txt)
- `user` - User accounts (username, password md5 + salt, mobile, last_ip, last_real_ip)
- `volume` - Book volumes (book_id, title, summary, cover)

列名统一为 snake_case：外键 `<表>_id`，布尔 `is_*`，计数 `*_count`。

## API Features

### Authentication
- JWT via `gin-jwt/v2`, middleware in `internal/middleware/jwt.go`
- Login: `POST /auth/login` with `{"username": "...", "password": "..."}` (password md5 + salt)
- Token timeout: 365 days; refresh via `GET /auth/refresh_token`
- Send `Authorization: Bearer <token>` on protected routes

### Routes
| Group | Auth | Examples |
| --- | --- | --- |
| `/` | No | `GET /` health check |
| `/search` | No | `GET /search` |
| `/options` | No | `GET /options` (page size + categories) |
| `/panic` | No | triggers a panic (recovery middleware test) |
| `/auth` | Mixed | `POST /auth/login`, `GET /auth/refresh_token`, `GET /auth/user`, `POST /auth/changepassword` |
| `/book` | Yes | books, download, finish, chapters, volumes |
| `/author` | Yes | `GET /author/index`, `GET /author/:author_id`, `POST /author/:author_id`, `DELETE /author/:author_id` |
| `/cnf` | Yes | `GET /cnf/categories`, `POST /cnf/category/:category_id`, `DELETE /cnf/category/:category_id` |

Responses are JSON: `{"code": 0, "msg": "...", "data": ...}`; `code` is 0 on success, otherwise the HTTP status (401/404/500).

### Additional Features
- CORS (allow origins from `[http] ACCESS_CONTROL_ALLOW_ORIGIN`)
- Optional GZIP and pprof (`[server] ENABLE_GZIP`, `ENABLE_PPROF`)
- Structured logging with file rotation
- Internationalization (i18n)
- Graceful shutdown on SIGINT/SIGTERM

## --daemon (self-daemonization)

`./book api --daemon` (or `-d`) detaches the server from the terminal; no `nohup` needed.

How it works (`internal/daemon`):
1. `book.go` `main` checks `daemon.Requested(os.Args[1:])` **before** CLI init, so the parent never opens the database or binds the port.
2. `daemon.Daemonize()` re-executes `os.Executable()` with the same args, sets `BOOK_DAEMONIZED=1`, redirects stdin to `/dev/null` and stdout/stderr to `nohup.out` (in the current working directory), and sets `SysProcAttr{Setsid: true}` so the child becomes a session leader without a controlling terminal.
3. The parent prints `book 已在后台启动，PID: <pid>，输出写入 nohup.out` and exits 0; the child keeps running in the background.
4. On the child, `daemon.IsChild()` is true, so daemonization is skipped and the normal CLI path runs.

Notes:
- `--daemon` only works with the `api` subcommand and must follow it (`book api --daemon`); `book --daemon api` is not supported.
- `nohup.out` is never rotated — program logs go to `<WORK DIR>/log/` with rotation, so trim `nohup.out` periodically or use launchd/systemd instead.
- Stop with `kill -15 <pid>` (the program shuts the HTTP server down gracefully); find the PID with `pgrep -f 'book api'`.
- `internal/daemon/daemon_test.go` covers flag detection and the env marker; an end-to-end check is `./book api --daemon` followed by `pgrep -f 'book api'`.

## Development Notes

### Code Style
- Mixed Chinese and English comments
- Uses `internal/` package for private implementation
- Error handling with `github.com/pkg/errors`
- Structured logging with Zap

### Hot Reload
- Configured with Air (`.air.toml`)
- Builds to `./tmp/main` and runs it with the `api` argument
- Excludes test files, `tmp/`, `vendor/`, `testdata/`

### Testing
- Test files use `_test.go` suffix
- Located alongside source files
- Use `github.com/stretchr/testify` for assertions
- `go test ./...` covers all packages; tests use temporary directories and inline config, and the single database-backed case (`internal/route`) skips itself when no database is reachable

### CI/CD
- GitHub Actions workflow in `.github/workflows/go.yml`
- Builds with Go 1.20 on Ubuntu (`go build -tags sqlite_fts5 -o book-api`)
- Artifacts named `book-api`

## Common Issues

### Configuration File Not Found
When running `go run book.go`, the `conf/` directory is not copied next to the temporary binary, so the embedded default config is used but `custom/conf/app.ini` may be missing. Solutions:
1. Set `BOOK_WORK_DIR`: `BOOK_WORK_DIR=/path/to/work go run book.go api`
2. Use `-c` flag: `go run book.go -c /path/to/custom/conf/app.ini api`
3. Add `export BOOK_WORK_DIR=/path/to/work` to your shell profile

### Port Already in Use
Change `[server] HTTP_PORT`, or stop the old process: `pgrep -f 'book api'` then `kill -15 <pid>`.

### Stopping the Server
- Foreground: `CTRL+C`
- Background (`--daemon`/nohup): `kill -15 <pid>`
- Linux style: `ps -ef | grep book | grep -v grep | awk '{print $2}' | xargs kill -15`

### Breaking Change Caution
- Renaming/removing config keys breaks custom configs silently (missing keys fall back to defaults); update `conf/app.ini` comments and this file together.
- Changing `internal/daemon` semantics: the parent must exit only after the child has started, and `BOOK_DAEMONIZED` must stay in sync or the process will recursively spawn itself.

## Related Projects
- Frontend: [book-frontend](https://github.com/zgia/book-frontend.git)
- Inspired by: [gogs](https://github.com/gogs/gogs) architecture
