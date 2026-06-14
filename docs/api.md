# API Documentation

Base URL: `http://localhost:8080` (local) or via Kubernetes ingress.

## Authentication

No authentication in POC. In production, add JWT validation middleware.

## Endpoints

### Health

| Method | Path | Description |
|---|---|---|
| GET | `/health/live` | Liveness probe — always 200 if process is up |
| GET | `/health/ready` | Readiness probe — 200 only if DB is reachable |

**GET /health/ready response:**
```json
{
  "status": "ok"
}
```

### Products

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products` | List products (paginated) |
| GET | `/api/v1/products/{id}` | Get product by ID |
| PUT | `/api/v1/products/{id}` | Update product |
| DELETE | `/api/v1/products/{id}` | Delete product |

### Metrics

| Method | Path | Description |
|---|---|---|
| GET | `/metrics` | Prometheus metrics |

---

## POST /api/v1/products

Create a new product.

**Request Body:**
```json
{
  "name": "Widget Pro",
  "description": "A professional widget",
  "price": 29.99,
  "stock": 100
}
```

**Validation Rules:**
- `name`: required, 1-255 characters
- `price`: required, > 0
- `stock`: required, ≥ 0
- `description`: optional

**Response: 201 Created**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Widget Pro",
    "description": "A professional widget",
    "price": 29.99,
    "stock": 100,
    "created_at": "2026-06-14T10:00:00Z",
    "updated_at": "2026-06-14T10:00:00Z"
  }
}
```

**Error Responses:**
- `422 Unprocessable Entity` — validation failed
- `500 Internal Server Error` — unexpected error

---

## GET /api/v1/products

List all products with pagination.

**Query Parameters:**
| Parameter | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number (1-based) |
| `page_size` | int | 20 | Items per page (max 100) |

**Response: 200 OK**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Widget Pro",
      "price": 29.99,
      "stock": 100,
      "created_at": "2026-06-14T10:00:00Z",
      "updated_at": "2026-06-14T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 42
  }
}
```

---

## GET /api/v1/products/{id}

Get a single product by UUID.

**Path Parameters:**
- `id`: UUID v4

**Response: 200 OK**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Widget Pro",
    "description": "A professional widget",
    "price": 29.99,
    "stock": 100,
    "created_at": "2026-06-14T10:00:00Z",
    "updated_at": "2026-06-14T10:00:00Z"
  }
}
```

**Error Responses:**
- `404 Not Found` — product not found

---

## PUT /api/v1/products/{id}

Update an existing product. All fields must be provided (full update).

**Request Body:**
```json
{
  "name": "Widget Pro v2",
  "description": "Updated description",
  "price": 39.99,
  "stock": 50
}
```

**Response: 200 OK** — returns updated product

**Error Responses:**
- `404 Not Found` — product not found
- `422 Unprocessable Entity` — validation failed

---

## DELETE /api/v1/products/{id}

Delete a product.

**Response: 204 No Content**

**Error Responses:**
- `404 Not Found` — product not found

---

## Error Response Format

All errors follow this format:
```json
{
  "error": "human-readable message",
  "code": "optional_machine_code"
}
```

## Quick Test with curl

```bash
# Create
curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","price":9.99,"stock":10}' | jq

# List
curl -s http://localhost:8080/api/v1/products | jq

# Get by ID (replace with actual UUID)
curl -s http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000 | jq

# Delete
curl -s -X DELETE http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000
```
