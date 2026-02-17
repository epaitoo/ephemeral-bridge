# EphemeralBridge

A self-hosted API for sharing ephemeral texts and files. Texts are stored in PostgreSQL, files are uploaded to Cloudflare R2 and automatically deleted after expiry. Built with Go.

## Features

- **Text sharing** -- Create, read, update, and delete text snippets
- **File sharing** -- Upload files (up to 10MB each) to Cloudflare R2 with presigned download URLs
- **Auto-expiry** -- Files expire 24 hours after first download and are cleaned up automatically
- **Authentication** -- API key auth, session cookies, and optional Cloudflare Access integration
- **Rate limiting** -- Built-in rate limiter (10 req/s, 20 burst)
- **Auto migrations** -- Database migrations run automatically on startup

## Tech Stack

- **Language:** Go
- **Router:** chi
- **Database:** PostgreSQL (local via Docker, or Neon for production)
- **Object Storage:** Cloudflare R2 (S3-compatible)
- **Auth:** Cloudflare Access (optional), API key, session cookies
- **Migrations:** golang-migrate

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose (for local PostgreSQL)
- A Cloudflare R2 bucket
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for manual migration commands)

### Setup

1. Clone the repository:

```bash
git clone https://github.com/epaitoo/ephermalbridge.git
cd ephermalbridge
```

2. Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

3. Start the local PostgreSQL database:

```bash
make db-up
```

4. Run the application:

```bash
go run ./cmd/api
```

The server starts on the port specified in your `.env` file (default `7500`). Migrations run automatically on startup.

## Configuration

All configuration is done through environment variables. See [.env.example](.env.example) for the full list.

| Variable | Description |
|---|---|
| `PORT` | Server port (default: 7500) |
| `APP_ENV` | Environment: `development`, `staging`, or `production` |
| `DATABASE_URL` | PostgreSQL connection string |
| `PROD_DATABASE_URL` | Production database URL (used for prod migrations via Makefile) |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 access key |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 secret key |
| `R2_ACCOUNT_ID` | Cloudflare account ID |
| `R2_BUCKET_NAME` | R2 bucket name |
| `R2_TOKEN_VALUE` | R2 API token |
| `R2_S3_API` | R2 S3-compatible API endpoint |
| `CLEANUP_INTERVAL_MINUTES` | How often to run expired file cleanup (default: 1440 min) |
| `APP_API_KEY` | API key for bearer token authentication |
| `APP_ALLOWED_EMAIL` | Email allowed for Cloudflare Access auth |
| `APP_COOKIE_SECRET` | Secret for signing session cookies |
| `APP_SKIP_CLOUDFLARE_AUTH` | Set to `true` to skip Cloudflare Access verification (for local dev) |
| `CLOUDFLARE_TEAM_DOMAIN` | Cloudflare Access team domain |
| `CLOUDFLARE_AUDIENCE` | Cloudflare Access audience tag |

## API Endpoints

All endpoints under `/v1/texts` and `/v1/files` require authentication via either:
- `Authorization: Bearer <API_KEY>` header, or
- A valid session cookie

### Health Check

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/healthcheck` | Server status (public) |

### Authentication

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/auth/session` | Create a session (requires Cloudflare Access JWT) |
| `POST` | `/v1/auth/logout` | Clear session cookie |

### Texts

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/texts` | List all texts |
| `GET` | `/v1/texts/{id}` | Get a text by ID |
| `POST` | `/v1/texts` | Create a new text |
| `PATCH` | `/v1/texts/{id}` | Update a text |
| `DELETE` | `/v1/texts/{id}` | Delete a text |

### Files

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/files` | List all files |
| `GET` | `/v1/files/{id}` | Get file metadata |
| `POST` | `/v1/files` | Upload files (multipart form, field: `files`) |
| `GET` | `/v1/files/{id}/download` | Get a presigned download URL (valid 10 min) |
| `DELETE` | `/v1/files/{id}` | Delete a file |
| `POST` | `/v1/files/cleanup` | Manually trigger expired file cleanup |

## Makefile Commands

```
make help               Show all available targets
make db-up              Start local PostgreSQL via Docker
make db-down            Stop local PostgreSQL
make migrate-up         Run migrations (dev database)
make migrate-down       Rollback last migration (dev database)
make migrate-create     Create a new migration (usage: make migrate-create name=add_users_table)
make migrate-force      Force set migration version (usage: make migrate-force version=1)
make migrate-version    Show current migration version (dev database)
make migrate-up-prod    Run migrations (production database)
make migrate-down-prod  Rollback last migration (production database)
make migrate-version-prod  Show migration version (production database)
```

## Project Structure

```
cmd/api/            Application entrypoint, HTTP handlers, routes, server
internal/
  auth/             Cloudflare Access JWT verification
  config/           Auth configuration loader
  data/             Database connection, migrations, models, R2 client
  middleware/       Auth, rate limiting, request logging, panic recovery
  upload/           R2 storage, file upload coordinator, cleanup scheduler
migrations/         SQL migration files
```

## License

MIT