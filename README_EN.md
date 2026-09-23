# Go Backend Frame · Server API

A Go foundation framework for admin systems and client-facing APIs. Built on Gin, GORM, MySQL, Redis and JWT, it provides dual service entrypoints, RBAC permissions, dynamic menus, login sessions, operation auditing, file uploads and common business channel configuration.

> The project is still evolving. Issues and pull requests are welcome.

| Website | Admin Client | API Server | API Docs |
| --- | --- | --- | --- |
| [Website](https://www.tutudati.com/) | [Admin source](https://gitee.com/open-source-project-open/go_basic_frame_admin) | [API source](https://gitee.com/open-source-project-open/go_basic_frame_api) | [API docs](https://s.apifox.cn/a42d392b-c5e9-4b75-8f54-e1c26339b262) |

[中文文档](README.md)

## Features

- Admin API and client API start independently with separate routers and can be deployed separately.
- JWT + Redis sessions with logout, password-change invalidation and admin forced offline.
- Menus, roles and departments are tree structures; admin side enforces API-level RBAC.
- Unified response envelope, parameter validation, pagination, database error and password handling.
- Automatic admin operation logs with sensitive field masking; log queries do not write logs themselves.
- Local, Aliyun OSS, Tencent COS, Qiniu Kodo and MinIO storage backends.
- Business tables store only relative file paths; URLs are resolved from the current storage config.
- Built-in SMS, WeChat, payment, storage and platform configuration modules.
- Startup self-check, HTTP timeouts and graceful shutdown.
- All tables and columns use `utf8mb4_general_ci`.
- Ships with an admin OpenAPI document, front-end/back-end code standards and a reusable code-standards skill.

## Tech Stack

| Category | Technology |
| --- | --- |
| Language | Go 1.26 |
| Web | Gin 1.10 |
| ORM | GORM 1.31 + MySQL driver |
| Cache | Redis 5+ |
| Auth | JWT v5 + Redis |
| CLI | Cobra |
| Config | YAML |
| Password | bcrypt |
| Storage | Local / Aliyun OSS / Tencent COS / Qiniu / MinIO |

## Project Structure

```text
server_api/
├── cmd/                       # CLI commands and service lifecycle
├── config/                    # Config structs and loading
├── internal/
│   ├── admin/                 # Admin business modules
│   │   ├── controller/        # Parameter binding and unified responses
│   │   ├── logic/             # Business logic
│   │   ├── middleware/        # Permission and operation log
│   │   ├── param/             # Request structs
│   │   ├── permission/        # RBAC computation and cache
│   │   └── resp/              # Response structs
│   ├── api/                   # Client business modules
│   └── common/                # Shared business capabilities
│       ├── app/               # MySQL/Redis init and checks
│       ├── auth/              # Sessions and JWT
│       ├── enums/             # Business enums
│       ├── middleware/        # Common middlewares
│       ├── model/             # GORM models
│       └── upload/            # Uploads and URL resolution
├── pkg/                       # Business-agnostic utilities
├── router/                    # admin/api route entrypoints
├── sql/                       # Incremental database scripts
├── uploads/                   # Local upload directory, not committed
├── docs/                      # Documentation
│   ├── admin_openapi.yaml     # Admin OpenAPI 3.0 document, Apifox-ready
│   ├── CODE_STYLE.md          # Full front-end/back-end code standards
│   └── code-standards/        # Code standards skill for AI coding assistants
│       ├── SKILL.md           # Skill entry: scenarios, order, cheat sheet
│       └── references/        # backend-go.md / frontend-vue.md details
├── config.example.yaml        # Configuration template
├── go.mod
└── main.go
```

## Requirements

- Go 1.26 or higher
- MySQL 5.7+ or MySQL 8.0+
- Redis 5+

## Quick Start

### 1. Clone and install dependencies

```bash
git clone https://gitee.com/open-source-project-open/go_basic_frame_api.git
cd go_basic_frame_api
go mod download
```

### 2. Initialize the database

`AutoMigrate` is not used; the schema must be maintained through SQL scripts.

`sql/` contains incremental scripts, for example:

```bash
mysql -uroot -p your_database < sql/platform_config.sql
```

The baseline script should create every table declared in `internal/common/model`. The service runs a table integrity check at startup and refuses to start when tables are missing or `sys_user` has no initial account.

> A complete `schema.sql` is not published in this repository yet. Before the first public release, provide a repeatable bootstrap script and keep further changes as separate incremental scripts.

### 3. Create the configuration file

```bash
cp config.example.yaml config.yaml
```

Update the MySQL, Redis and JWT sections. Always replace the JWT secret in production:

```bash
openssl rand -hex 32
```

Key options:

| Option | Description | Example |
| --- | --- | --- |
| `server.admin_addr` | Admin API address | `:8001` |
| `server.api_addr` | Client API address | `:8002` |
| `mysql.*` | MySQL connection and pool | See template |
| `redis.*` | Redis address, password, DB | See template |
| `jwt.secret` | JWT signing secret | Random strong secret |
| `jwt.expire_hours` | Session lifetime | `24` |
| `log.level` | Log level | `info` |

`config.yaml` may contain database passwords and secrets; never commit real production configuration to a public repository.

### 4. Run the services

```bash
# Admin API, listens on :8001 by default
go run . service admin -c config.yaml

# Client API, listens on :8002 by default
go run . service api -c config.yaml
```

### 5. Build

```bash
go build -o bin/server_api .

./bin/server_api service admin -c config.yaml
./bin/server_api service api -c config.yaml
```

Other commands:

```bash
go run . version
go run . help
```

## API Conventions

### Unified response

All endpoints return a unified JSON envelope from `pkg/response`:

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

Common business codes:

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `400` | Invalid parameters |
| `401` | Not logged in or session expired |
| `403` | No permission for this endpoint |
| `500` | Business or server error |
| `503` | Dependency temporarily unavailable |

### Authentication

Protected endpoints use a bearer token:

```http
Authorization: Bearer <token>
```

- Admin: `Auth → Permission → OperationLog`.
- Client: `Auth`.
- Super admins bypass endpoint matching but still need a valid session.
- Regular admin permissions come from the `METHOD:/route` entries of menus bound to their roles.

### Service Entrypoints

| Service | Prefix | Default port | Description |
| --- | --- | --- | --- |
| Admin | `/admin` | `8001` | RBAC, configuration and system management |
| Client | `/api` | `8002` | Client login, profile and uploads |
| Local files | `/files/*filepath` | Both | Local storage backend only |

Routes are the source of truth; see [router/admin.go](router/admin.go) and [router/api.go](router/api.go).

The admin API also ships an OpenAPI 3.0 document: [docs/admin_openapi.yaml](docs/admin_openapi.yaml). Import it into Apifox, Postman or Swagger UI and change `servers.url` to your address.

## Database Conventions

- Every table and column must have a comment.
- Character columns use `utf8mb4_general_ci`.
- Enum values start at `1` and live in `internal/common/enums`.
- GORM models must carry `gorm`, `json` tags and field comments.
- Relations are expressed in response structs, never coupled into models.
- Prefer GORM `Preload` for related queries.
- File columns store relative paths only; `sys_upload_file` keeps both the relative path and the full URL.
- Schema changes ship as incremental scripts in `sql/`; runtime auto-migration is forbidden.

## Code Standards

The full standards live in [docs/CODE_STYLE.md](docs/CODE_STYLE.md) and cover directory layout, layer responsibilities, naming, templates and pre-submit checklists for Go and Vue/TypeScript.

Core rules:

- `controller` only binds parameters, calls logic and writes the response.
- Business rules, transactions and context handling belong to `logic`; `logic` methods accept `*gin.Context`.
- `logic` never returns models directly; convert to `resp` structs.
- Split `param` and `resp` by feature; a single `param.go`/`resp.go` is not allowed.
- Shared capabilities go to `internal/common/<feature>`.
- Business-agnostic helpers go to `pkg/<feature>`.
- Always respond through `OK/Fail` in `pkg/response` with HTTP 200 and business codes for errors.
- New endpoints must consider authentication, input validation, sensitive data masking and concurrent write safety.
- `AutoMigrate` is forbidden; schema changes ship as `sql/` scripts.
- Update `docs/admin_openapi.yaml` on API changes; update `config/config.go`, `config.example.yaml` and the config table above on config changes.

### Code Standards Skill

[docs/code-standards](docs/code-standards) is the AI-assistant version of the same standards:

```text
docs/code-standards/
├── SKILL.md                   # Scenarios, backend/frontend order, cheat sheet
└── references/
    ├── backend-go.md          # Layer responsibilities, templates, pitfalls
    └── frontend-vue.md        # Types/API/page templates, style and checks
```

To use it in CodeBuddy, Claude Code or similar tools:

```bash
mkdir -p .codebuddy/skills
cp -r docs/code-standards .codebuddy/skills/go-frame-code-standards
```

Once copied, the standards load automatically when writing, modifying or reviewing `server_api` and `admin_client` code; you can also mention `go-frame-code-standards` explicitly. Keep `docs/CODE_STYLE.md` and `docs/code-standards/` in sync when the standards change.

## Development and Checks

```bash
# Format
go fmt ./...

# Test
go test ./...

# Vet
go vet ./...
```

Do not commit temporary test files, build artifacts, logs, uploads, real configuration or secrets.

## Deployment

- Terminate HTTPS at a reverse proxy and limit upload size and request rate.
- Run admin and client APIs as separate processes and domains.
- Restrict MySQL and Redis to trusted networks.
- Tune connection pools and HTTP timeouts for real traffic.
- Back up the database and object storage regularly and verify restores.
- Add structured logging, metrics and alerting in production.

## Contributing

1. Fork the repository and create a feature branch from the main branch.
2. Keep changes focused and follow [docs/CODE_STYLE.md](docs/CODE_STYLE.md) (the [code standards skill](docs/code-standards) helps).
3. Run `gofmt`, `go test ./...` and `go vet ./...` before committing.
4. Describe the purpose, database impact, compatibility and verification in the pull request.

Do not disclose security vulnerabilities in public issues; report them through the maintainer's private channel.

## Screenshots

### Dashboard

![Dashboard](./images/v1_pre/Snipaste_2026-09-23_22-59-04.png)

### Storage Configuration

![Storage Configuration](./images/v1_pre/Snipaste_2026-09-23_22-59-23.png)

### Platform Configuration

![Platform Configuration](./images/v1_pre/Snipaste_2026-09-23_22-59-36.png)

### Menu Management

![Menu Management](./images/v1_pre/Snipaste_2026-09-23_22-59-58.png)

### User Management

![User Management](./images/v1_pre/Snipaste_2026-09-23_23-00-09.png)

### Role Management

![Role Management](./images/v1_pre/Snipaste_2026-09-23_23-00-23.png)

### Operation Logs

![Operation Logs](./images/v1_pre/Snipaste_2026-09-23_23-00-36.png)

## WeChat

Scan the QR code below to discuss usage, ideas or contributions:

<p align="left">
  <img src="./images/wechat.png" width="280" height="350" alt="WeChat QR code" />
</p>

## License

Released under the [Apache License 2.0](LICENSE).
