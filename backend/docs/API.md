# CodeXray Observability Service API Documentation

## Overview

The CodeXray Observability Service provides REST APIs for system monitoring, alerting, and log analysis.

**Base URL:** `http://localhost:8080/api/v1`

## Authentication

Most endpoints require authentication using a session token obtained from the login endpoint.

**Header:** `Authorization: Bearer <token>` or `Authorization: <token>`

## Endpoints

### Health Check

#### GET /health
Check if the service is running.

**Response:**
```json
{
  "status": "healthy",
  "message": "CodeXray Observability Service is running"
}
```

### Authentication

#### POST /api/v1/auth/register
Register a new user.

**Request Body:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### POST /api/v1/auth/login
Authenticate a user and get a session token.

**Request Body:**
```json
{
  "username": "john_doe",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "token": "abc123def456...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

#### POST /api/v1/auth/validate
Validate a session token.

**Request Body:**
```json
{
  "token": "abc123def456..."
}
```

**Response:**
```json
{
  "valid": true,
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

#### POST /api/v1/auth/logout
Logout and invalidate session token.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "message": "Logout successful"
}
```

### Log Analysis

#### GET /api/v1/logs/analyze?file=<path>
Analyze a log file and return statistics.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `file` (required): Path to the log file

**Response:**
```json
{
  "message": "Log analysis completed",
  "stats": {
    "level_counts": {
      "INFO": 8,
      "WARN": 4,
      "ERROR": 6,
      "DEBUG": 2
    },
    "top_errors": [
      {
        "message": "Failed to connect to external API: connection timeout",
        "count": 3
      },
      {
        "message": "Database query failed: table 'users' doesn't exist",
        "count": 2
      }
    ],
    "total_entries": 20
  }
}
```

### Metrics

#### GET /api/v1/metrics/current
Get current system metrics.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "message": "Current metrics retrieved",
  "metrics": {
    "cpu_usage": 45.2,
    "memory_usage": 68.7,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

#### GET /api/v1/metrics/history/:type?limit=<n>
Get historical metrics for a specific type.

**Headers:** `Authorization: Bearer <token>`

**Path Parameters:**
- `type`: Metric type (`cpu_usage` or `memory_usage`)

**Query Parameters:**
- `limit` (optional): Number of records to return (default: 100)

**Response:**
```json
{
  "message": "Metric history retrieved",
  "history": [
    {
      "id": 1,
      "type": "cpu_usage",
      "value": 45.2,
      "unit": "%",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Alerts

#### GET /api/v1/alerts?status=<status>&limit=<n>
Get alerts with optional filtering.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `status` (optional): Filter by status (`active` or `resolved`)
- `limit` (optional): Number of records to return (default: 50)

**Response:**
```json
{
  "message": "Alerts retrieved",
  "alerts": [
    {
      "id": 1,
      "type": "cpu_usage",
      "message": "High CPU usage detected: 85.2% (threshold: 80.0%)",
      "value": 85.2,
      "threshold": 80.0,
      "severity": "high",
      "status": "active",
      "triggered_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### POST /api/v1/alerts
Manually create an alert (for testing).

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "type": "cpu_usage",
  "value": 85.2,
  "threshold": 80.0
}
```

**Response:**
```json
{
  "message": "Alert created",
  "alert": {
    "id": 1,
    "type": "cpu_usage",
    "message": "High CPU usage detected: 85.2% (threshold: 80.0%)",
    "value": 85.2,
    "threshold": 80.0,
    "severity": "high",
    "status": "active",
    "triggered_at": "2024-01-15T10:30:00Z"
  }
}
```

#### PUT /api/v1/alerts/:id/resolve
Resolve an alert.

**Headers:** `Authorization: Bearer <token>`

**Path Parameters:**
- `id`: Alert ID

**Response:**
```json
{
  "message": "Alert resolved"
}
```

### Summary Report

#### GET /api/v1/summary?limit=<n>
Get comprehensive system summary.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `limit` (optional): Number of recent alerts to include (default: 10)

**Response:**
```json
{
  "message": "Summary retrieved",
  "summary": {
    "current_metrics": {
      "cpu_usage": 45.2,
      "memory_usage": 68.7,
      "timestamp": "2024-01-15T10:30:00Z"
    },
    "alerts": {
      "total_alerts": 15,
      "active_alerts": 2,
      "resolved_alerts": 13,
      "alerts_by_type": {
        "cpu_usage": 8,
        "memory_usage": 7
      },
      "alerts_by_severity": {
        "low": 3,
        "medium": 6,
        "high": 4,
        "critical": 2
      },
      "recent_alerts": [...]
    },
    "metric_averages": {
      "cpu": {
        "type": "cpu_usage",
        "average": 52.3,
        "min": 23.1,
        "max": 89.4,
        "count": 10
      },
      "memory": {
        "type": "memory_usage",
        "average": 71.8,
        "min": 45.2,
        "max": 92.1,
        "count": 10
      }
    }
  }
}
```

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid token)
- `404` - Not Found
- `409` - Conflict (e.g., user already exists)
- `500` - Internal Server Error

## Rate Limiting

Currently, no rate limiting is implemented, but it's recommended for production use.

## CORS

CORS is enabled for all origins in development. Configure appropriately for production.