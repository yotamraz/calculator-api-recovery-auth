# Calculator API

A simple calculator REST API with authentication, built with FastAPI, SQLModel, and SQLite.

## Setup

```bash
pip install -e .[dev]
```

## Run

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

Or after installing:

```bash
calculator-api
```

The server runs on `http://localhost:8000`.

## Configuration

All settings are configurable via environment variables with the `CALC_API_` prefix. You can also place them in a `.env` file in the project root.

| Environment Variable | Description | Default |
|---|---|---|
| `CALC_API_DATABASE_URL` | SQLite database URL | `sqlite:///calculator.db` |
| `CALC_API_JWT_SECRET_KEY` | Secret used to sign JWT tokens | `dev-secret-key-change-me-in-production` |
| `CALC_API_JWT_ALGORITHM` | JWT signing algorithm | `HS256` |
| `CALC_API_ACCESS_TOKEN_EXPIRE_MINUTES` | Token expiry time in minutes | `30` |

> **Important:** Always set `CALC_API_JWT_SECRET_KEY` to a strong, unique value in production.

## Authentication

All endpoints (except `/health`) require a JWT Bearer token. The auth flow uses the standard OAuth2 password grant.

### Register a user

```bash
curl -X POST http://localhost:8000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret123"}'
```

Response:

```json
{"id": 1, "username": "alice", "created_at": "2026-02-16T14:56:53.344519"}
```

### Obtain a token

```bash
curl -X POST http://localhost:8000/auth/token \
  -d "username=alice&password=secret123"
```

Response:

```json
{"access_token": "eyJhbGciOi...", "token_type": "bearer"}
```

### Use the token

Pass the token in the `Authorization` header for all subsequent requests:

```bash
curl http://localhost:8000/calculations \
  -H "Authorization: Bearer <token>"
```

## Endpoints

### Health Check

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/health` | No | Service health status |

### Auth

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Register a new user |
| POST | `/auth/token` | No | Login and get a JWT token |

### Calculator Operations

All calculator endpoints accept POST with JSON body `{"a": <number>, "b": <number>}` and require a valid Bearer token.

| Endpoint | Description |
|---|---|
| `/add` | Add two numbers |
| `/subtract` | Subtract b from a |
| `/multiply` | Multiply two numbers |
| `/divide` | Divide a by b |

### Calculation History (CRUD)

Store calculations in a SQLite database. All endpoints require a valid Bearer token.

| Method | Endpoint | Description |
|---|---|---|
| POST | `/calculations` | Create & store a calculation |
| GET | `/calculations` | List all stored calculations |
| GET | `/calculations/{id}` | Get a specific calculation |
| DELETE | `/calculations/{id}` | Delete a calculation |

## Examples

**Quick calculation (authenticated):**

```bash
curl -X POST http://localhost:8000/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"a": 5, "b": 3}'
```

Response: `{"result": 8.0}`

**Save a calculation to the database:**

```bash
curl -X POST http://localhost:8000/calculations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"operation": "mul", "a": 7, "b": 6}'
```

Response: `{"operation": "mul", "a": 7.0, "b": 6.0, "result": 42.0, "id": 1, "created_at": "..."}`

**List all saved calculations:**

```bash
curl http://localhost:8000/calculations \
  -H "Authorization: Bearer <token>"
```

**Delete a calculation:**

```bash
curl -X DELETE http://localhost:8000/calculations/1 \
  -H "Authorization: Bearer <token>"
```

## Testing

Run the test suite:

```bash
pip install -e .[dev]
pytest
```

## API Docs

FastAPI auto-generates docs at:
- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc
