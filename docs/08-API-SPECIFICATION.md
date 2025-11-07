# 08-API-SPECIFICATION.md

```markdown
# 08-API-SPECIFICATION.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** API Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md), 
[02-DATA-MODEL.md](02-DATA-MODEL.md)

---

## 📖 Table of Contents

1. [API Overview](#api-overview)
2. [Authentication](#authentication)
3. [Task Endpoints](#task-endpoints)
4. [Agent Endpoints](#agent-endpoints)
5. [Optimization Endpoints](#optimization-endpoints)
6. [Consensus Endpoints](#consensus-endpoints)
7. [WebSocket API](#websocket-api)
8. [Error Responses](#error-responses)
9. [Data Transfer Objects](#data-transfer-objects)

---

## 🌐 API Overview

### Base URL

**Development:**
```
http://localhost/api/v1
```

**Production (via Caddy):**
```
https://afs.example.com/api/v1
```

### API Conventions

**RESTful Principles:**
- Resource-based URLs
- HTTP methods semantically correct (GET, POST, PUT, DELETE)
- Plural nouns for collections (`/tasks`, `/agents`)
- Singular for single resource (`/tasks/{id}`)

**Request/Response Format:**
- Content-Type: `application/json`
- Character encoding: UTF-8
- Date format: ISO 8601 (`2024-01-15T10:30:00Z`)
- Numbers: No quotes (native JSON numbers)

**Versioning:**
- URL-based versioning (`/v1/`)
- Future versions: `/v2/` (backward compatible)

**HTTP Status Codes:**
- 200 OK: Successful GET, PUT
- 201 Created: Successful POST
- 204 No Content: Successful DELETE
- 400 Bad Request: Validation error
- 404 Not Found: Resource doesn't exist
- 409 Conflict: Duplicate resource
- 500 Internal Server Error: Server error

**Pagination (Lists):**
- Query params: `?limit=20&offset=0`
- Default limit: 20
- Max limit: 100
- Response includes metadata (total count, has_more)

**Filtering (Lists):**
- Query params for filtering: `?status=completed`
- Multiple filters: `?status=completed&type=query_optimization`
- Case-insensitive where applicable

**Sorting (Lists):**
- Query param: `?sort=-created_at` (descending)
- Default: `-created_at` (newest first)
- Prefix `-` for descending, no prefix for ascending

---

## 🔐 Authentication

### Current Implementation

**Development Phase:**
- No authentication (open API)
- Focus on functionality over security
- Suitable for local development only

**Production Phase (Future):**
- API key authentication
- Header: `Authorization: Bearer {api_key}`
- Rate limiting per API key
- User-based resource isolation

**Challenge Submission:**
- Demo credentials provided to judges
- Read-only API key for testing
- Limited rate (100 requests/minute)

---

## 📋 Task Endpoints

### POST /tasks

**Purpose:** Create new optimization task

**Request:**

**Headers:**
```
Content-Type: application/json
```

**Body:**
```json
{
  "type": "query_optimization",
  "description": "Optimize monthly revenue report",
  "target_query": "SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email ORDER BY revenue DESC",
  "metadata": {
    "priority": "high",
    "complexity_hint": "medium",
    "target_tables": ["orders", "users"],
    "user_preferences": {
      "max_storage_mb": 50,
      "risk_tolerance": "low"
    },
    "scoring_weights": {
      "performance": 0.5,
      "storage": 0.2,
      "complexity": 0.2,
      "risk": 0.1
    }
  }
}
```

**Field Descriptions:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| type | string | Yes | Enum: query_optimization, schema_improvement, index_recommendation |
| description | string | No | Max 500 chars |
| target_query | string | Yes | Min 10 chars, valid SQL syntax |
| metadata | object | No | See metadata schema below |

**Metadata Schema:**

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| priority | string | No | "medium" | Enum: low, medium, high |
| complexity_hint | string | No | null | Enum: low, medium, high |
| target_tables | array | No | [] | Array of table names |
| user_preferences | object | No | {} | See preferences schema |
| scoring_weights | object | No | null | Must sum to 1.0 |

**Response (201 Created):**

```json
{
  "id": 123,
  "type": "query_optimization",
  "description": "Optimize monthly revenue report",
  "target_query": "SELECT u.email...",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": null,
  "metadata": {
    "priority": "high",
    "complexity_hint": "medium",
    "target_tables": ["orders", "users"]
  },
  "links": {
    "self": "/api/v1/tasks/123",
    "agents": "/api/v1/tasks/123/agents",
    "proposals": "/api/v1/tasks/123/proposals"
  }
}
```

**Validation Errors (400):**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "target_query": "Query cannot be empty",
      "type": "Invalid type: must be query_optimization, schema_improvement, or index_recommendation"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Business Logic:**
1. Validate request body
2. Create Task entity
3. Persist to database (status: pending)
4. Trigger async orchestration (goroutine)
5. Return task immediately (don't wait for completion)

---

### GET /tasks/{id}

**Purpose:** Get task details by ID

**Path Parameters:**
- `id` (integer): Task identifier

**Response (200 OK):**

```json
{
  "id": 123,
  "type": "query_optimization",
  "description": "Optimize monthly revenue report",
  "target_query": "SELECT u.email, SUM(o.total)...",
  "status": "in_progress",
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": null,
  "metadata": {
    "priority": "high",
    "current_step": "benchmarking",
    "agents_working": ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"]
  },
  "links": {
    "self": "/api/v1/tasks/123",
    "agents": "/api/v1/tasks/123/agents",
    "proposals": "/api/v1/tasks/123/proposals",
    "consensus": "/api/v1/tasks/123/consensus"
  }
}
```

**Status Field Values:**
- `pending`: Queued, not started
- `in_progress`: Agents actively working
- `completed`: Successfully finished
- `failed`: Error occurred

**Current Step (in metadata, optional):**
- `routing`: Selecting agents
- `fork_creation`: Creating forks
- `analysis`: Agents analyzing query
- `proposal`: Generating optimization proposals
- `benchmarking`: Running performance tests
- `consensus`: Evaluating proposals
- `applying`: Applying to main database
- `cleanup`: Destroying forks

**Not Found (404):**

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Task with id 123 not found",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

### GET /tasks

**Purpose:** List all tasks with filtering and pagination

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | integer | 20 | Results per page (max 100) |
| offset | integer | 0 | Skip N results |
| status | string | null | Filter by status |
| type | string | null | Filter by task type |
| sort | string | -created_at | Sort field (- for desc) |

**Example Request:**
```
GET /api/v1/tasks?status=completed&limit=10&offset=0
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": 125,
      "type": "query_optimization",
      "description": "Optimize user orders query",
      "status": "completed",
      "created_at": "2024-01-15T12:00:00Z",
      "completed_at": "2024-01-15T12:05:30Z"
    },
    {
      "id": 124,
      "type": "index_recommendation",
      "description": "Index recommendation for payments table",
      "status": "completed",
      "created_at": "2024-01-15T11:30:00Z",
      "completed_at": "2024-01-15T11:33:15Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 45,
    "has_more": true
  },
  "links": {
    "self": "/api/v1/tasks?status=completed&limit=10&offset=0",
    "next": "/api/v1/tasks?status=completed&limit=10&offset=10",
    "prev": null
  }
}
```

**Empty Result (200 OK):**

```json
{
  "data": [],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 0,
    "has_more": false
  }
}
```

---

### GET /tasks/{id}/agents

**Purpose:** Get agent executions for a task

**Response (200 OK):**

```json
{
  "task_id": 123,
  "agents": [
    {
      "id": 456,
      "agent_type": "gemini-2.5-pro",
      "fork_id": "afs-fork-gemini-2.5-pro-task123-1699901234",
      "status": "completed",
      "started_at": "2024-01-15T10:30:05Z",
      "completed_at": "2024-01-15T10:33:20Z",
      "duration_seconds": 195,
      "error_message": null
    },
    {
      "id": 457,
      "agent_type": "gemini-2.0-flash",
      "fork_id": "afs-fork-gemini-2.0-flash-task123-1699901235",
      "status": "completed",
      "started_at": "2024-01-15T10:30:05Z",
      "completed_at": "2024-01-15T10:33:45Z",
      "duration_seconds": 220,
      "error_message": null
    },
    {
      "id": 458,
      "agent_type": "gemini-2.5-flash",
      "fork_id": "afs-fork-gemini-2.5-flash-task123-1699901236",
      "status": "completed",
      "started_at": "2024-01-15T10:30:05Z",
      "completed_at": "2024-01-15T10:33:10Z",
      "duration_seconds": 185,
      "error_message": null
    }
  ]
}
```

**Agent Status Values:**
- `running`: Currently executing
- `completed`: Finished successfully
- `failed`: Encountered error

---

## 🤖 Agent Endpoints

### GET /agents

**Purpose:** List all available agents with their capabilities

**Response (200 OK):**

```json
{
  "agents": [
    {
      "type": "gemini-2.5-pro",
      "name": "gemini-2.5-pro",
      "status": "available",
      "specialization": {
        "strengths": [
          "Planning/QA",
          "Requirement disambiguation",
          "SQL/EXPLAIN review",
          "Consistency checks"
        ],
        "description": "Planner/QA for critical decisions",
        "preferred_task_types": ["query_optimization","schema_improvement"]
      },
      "metrics": {"active_tasks": 1,"total_tasks_completed": 45,"success_rate": 0.96,"avg_task_duration_seconds": 180}
    },
    {
      "type": "gemini-2.5-flash",
      "name": "gemini-2.5-flash",
      "status": "available",
      "specialization": {
        "strengths": ["SQL/code generation","Transformations","Benchmarks"],
        "description": "Low-latency code/SQL generation and testing",
        "preferred_task_types": ["query_optimization","schema_improvement"]
      },
      "metrics": {"active_tasks": 0,"total_tasks_completed": 38,"success_rate": 0.89,"avg_task_duration_seconds": 210}
    },
    {
      "type": "gemini-2.0-flash",
      "name": "gemini-2.0-flash",
      "status": "available",
      "specialization": {
        "strengths": ["Boilerplate","Refactors simples","Formatters"],
        "description": "High-throughput low-risk tasks",
        "preferred_task_types": ["index_recommendation","query_optimization"]
      },
      "metrics": {"active_tasks": 2,"total_tasks_completed": 42,"success_rate": 0.93,"avg_task_duration_seconds": 195}
    }
  ]
}
```

**Agent Status Values:**
- `available`: Can accept new tasks
- `busy`: At capacity (3 concurrent tasks)
- `unavailable`: Offline or error state

---

### GET /agents/{type}/status

**Purpose:** Get detailed status of specific agent

**Path Parameters:**
- `type` (string): Agent type (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)

**Response (200 OK):**

```json
{
  "type": "gemini-2.5-pro",
  "status": "available",
  "current_tasks": [
    {
      "task_id": 123,
      "started_at": "2024-01-15T10:30:00Z",
      "current_step": "benchmarking"
    }
  ],
  "capacity": {
    "max_concurrent": 3,
    "current_active": 1,
    "available_slots": 2
  },
  "recent_performance": {
    "last_10_tasks": {
      "success_rate": 1.0,
      "avg_duration_seconds": 175,
      "avg_improvement_pct": 78.5
    }
  }
}
```

---

## 🔧 Optimization Endpoints

### GET /tasks/{id}/proposals

**Purpose:** Get all optimization proposals for a task

**Response (200 OK):**

```json
{
  "task_id": 123,
  "proposals": [
    {
      "id": 45,
      "agent_type": "gemini-2.5-pro",
      "agent_execution_id": 456,
      "proposal_type": "partial_index",
      "sql_commands": [
        "CREATE INDEX idx_orders_user_completed ON orders(user_id, status) WHERE status = 'completed'",
        "ANALYZE orders"
      ],
      "rationale": "Partial index targets only completed orders (75% of data), reducing index size. Covers both JOIN condition (user_id) and filter (status), enabling index-only scans.",
      "estimated_impact": {
        "query_time_improvement": 80.0,
        "storage_overhead_mb": 12.0,
        "complexity": "low",
        "risk": "low"
      },
      "benchmark_summary": {
        "avg_improvement_pct": 82.6,
        "min_improvement_pct": 80.0,
        "max_improvement_pct": 85.0,
        "storage_overhead_mb": 12.0,
        "test_count": 4
      },
      "created_at": "2024-01-15T10:31:30Z"
    },
    {
      "id": 46,
      "agent_type": "gemini-2.5-flash",
      "agent_execution_id": 458,
      "proposal_type": "materialized_view",
      "sql_commands": [
        "CREATE MATERIALIZED VIEW mv_user_revenue AS SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email",
        "CREATE UNIQUE INDEX idx_mv_user_revenue_email ON mv_user_revenue(email)",
        "CREATE INDEX idx_mv_user_revenue_revenue ON mv_user_revenue(revenue DESC)"
      ],
      "rationale": "Materialized view pre-computes aggregation, turning complex query into simple table scan. Best for frequently-read, infrequently-updated data. Requires concurrent refresh strategy to minimize locking.",
      "estimated_impact": {
        "query_time_improvement": 90.0,
        "storage_overhead_mb": 80.0,
        "complexity": "medium",
        "risk": "medium"
      },
      "benchmark_summary": {
        "avg_improvement_pct": 93.5,
        "min_improvement_pct": 90.0,
        "max_improvement_pct": 95.0,
        "storage_overhead_mb": 80.0,
        "test_count": 4
      },
      "created_at": "2024-01-15T10:31:45Z"
    },
    {
      "id": 47,
      "agent_type": "gemini-2.0-flash",
      "agent_execution_id": 457,
      "proposal_type": "partitioning",
      "sql_commands": [
        "CREATE TABLE orders_partitioned (LIKE orders INCLUDING ALL) PARTITION BY LIST (status)",
        "CREATE TABLE orders_completed PARTITION OF orders_partitioned FOR VALUES IN ('completed')",
        "CREATE TABLE orders_other PARTITION OF orders_partitioned FOR VALUES IN ('pending', 'processing')"
      ],
      "rationale": "Partitioning by status enables partition pruning, automatically excluding 'pending' and 'processing' partitions from query scans. Scales better for future growth.",
      "estimated_impact": {
        "query_time_improvement": 60.0,
        "storage_overhead_mb": 40.0,
        "complexity": "high",
        "risk": "medium"
      },
      "benchmark_summary": {
        "avg_improvement_pct": 61.5,
        "min_improvement_pct": 50.0,
        "max_improvement_pct": 70.0,
        "storage_overhead_mb": 40.0,
        "test_count": 4
      },
      "created_at": "2024-01-15T10:32:00Z"
    }
  ],
  "count": 3
}
```

**No Proposals Yet (200 OK):**

```json
{
  "task_id": 123,
  "proposals": [],
  "count": 0,
  "message": "Agents are still analyzing the task"
}
```

---

### GET /proposals/{id}/benchmarks

**Purpose:** Get detailed benchmark results for a proposal

**Response (200 OK):**

```json
{
  "proposal_id": 45,
  "benchmarks": [
    {
      "id": 789,
      "query_name": "baseline_test1",
      "query_executed": "SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email ORDER BY revenue DESC",
      "execution_time_ms": 2300.45,
      "rows_returned": 1000,
      "explain_plan": {
        "planning_time_ms": 0.5,
        "execution_time_ms": 2299.95,
        "plan_type": "Hash Join",
        "index_name": null,
        "actual_rows": 1000,
        "buffers": {
          "shared_hit": 250,
          "shared_read": 50,
          "cache_hit_rate": 0.83
        }
      },
      "storage_impact_mb": null,
      "created_at": "2024-01-15T10:32:10Z"
    },
    {
      "id": 790,
      "query_name": "optimized_test1",
      "query_executed": "SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email ORDER BY revenue DESC",
      "execution_time_ms": 450.23,
      "rows_returned": 1000,
      "explain_plan": {
        "planning_time_ms": 0.6,
        "execution_time_ms": 449.63,
        "plan_type": "Index Scan",
        "index_name": "idx_orders_user_completed",
        "actual_rows": 1000,
        "filter_removed_rows": 0,
        "buffers": {
          "shared_hit": 180,
          "shared_read": 5,
          "cache_hit_rate": 0.97
        }
      },
      "storage_impact_mb": 12.0,
      "improvement_pct": 80.4,
      "created_at": "2024-01-15T10:32:45Z"
    },
    {
      "id": 791,
      "query_name": "baseline_test2",
      "query_executed": "SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email ORDER BY revenue DESC LIMIT 10",
      "execution_time_ms": 800.0,
      "rows_returned": 10,
      "created_at": "2024-01-15T10:32:15Z"
    },
    {
      "id": 792,
      "query_name": "optimized_test2",
      "query_executed": "SELECT u.email, SUM(o.total) as revenue FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.email ORDER BY revenue DESC LIMIT 10",
      "execution_time_ms": 120.0,
      "rows_returned": 10,
      "improvement_pct": 85.0,
      "created_at": "2024-01-15T10:32:50Z"
    }
  ],
  "summary": {
    "total_tests": 8,
    "avg_improvement_pct": 82.6,
    "storage_overhead_mb": 12.0,
    "execution_time_range": "120ms - 450ms"
  }
}
```

---

## ⚖️ Consensus Endpoints

### GET /tasks/{id}/consensus

**Purpose:** Get consensus decision for a task

**Response (200 OK - Decision Made):**

```json
{
  "task_id": 123,
  "winning_proposal_id": 45,
  "winner_agent": "gemini-2.5-pro",
  "applied_to_main": true,
  "scores": {
    "gemini-2.5-pro": {
      "proposal_id": 45,
      "performance_score": 95.0,
      "storage_score": 95.0,
      "complexity_score": 85.0,
      "risk_score": 95.0,
      "weighted_total": 93.0,
      "rank": 1,
      "improvement_pct": 82.6,
      "storage_overhead_mb": 12.0
    },
    "gemini-2.5-flash": {
      "proposal_id": 46,
      "performance_score": 100.0,
      "storage_score": 60.0,
      "complexity_score": 50.0,
      "risk_score": 65.0,
      "weighted_total": 78.5,
      "rank": 2,
      "improvement_pct": 93.5,
      "storage_overhead_mb": 80.0
    },
    "gemini-2.0-flash": {
      "proposal_id": 47,
      "performance_score": 76.0,
      "storage_score": 75.0,
      "complexity_score": 35.0,
      "risk_score": 65.0,
      "weighted_total": 66.5,
      "rank": 3,
      "improvement_pct": 61.5,
      "storage_overhead_mb": 40.0
    }
  },
  "decision_rationale": "gemini-2.5-pro partial index proposal selected as optimal solution.\n\nPerformance: 82.6% average improvement (2.30s → 0.40s)\n\nKey Strengths:\n- Minimal storage overhead (12MB vs 80MB for materialized view)\n- Low operational complexity with easy rollback path (simple DROP INDEX)\n- Consistent performance across all test scenarios\n\nTrade-off Analysis:\n- While gemini-2.5-flash materialized view achieved highest raw performance (93.5% improvement), the 80MB storage overhead and refresh maintenance complexity made it less balanced. gemini-2.5-pro offers 82.6% improvement with negligible overhead and zero maintenance.\n\nRunner-up: gemini-2.5-flash materialized_view proposal (78.5 points)\nWhile it achieved 93.5% improvement, the storage overhead (80MB) and ongoing refresh maintenance requirements reduced its overall score.",
  "created_at": "2024-01-15T10:34:00Z"
}
```

**Response (200 OK - No Decision Yet):**

```json
{
  "task_id": 123,
  "winning_proposal_id": null,
  "applied_to_main": false,
  "message": "Consensus not yet reached - agents still working",
  "current_status": "benchmarking"
}
```

**Response (200 OK - No Winner):**

```json
{
  "task_id": 123,
  "winning_proposal_id": null,
  "applied_to_main": false,
  "scores": {
    "gemini-2.5-pro": {
      "weighted_total": 25.0,
      "improvement_pct": 8.0
    },
    "gemini-2.5-flash": {
      "weighted_total": 22.0,
      "improvement_pct": 6.5
    }
  },
  "decision_rationale": "No significant optimization found. All proposals showed <10% improvement, suggesting query is already well-optimized or bottleneck is external to database.",
  "created_at": "2024-01-15T10:34:00Z"
}
```

---

## 🔌 WebSocket API

### Connection

**Endpoint:**
```
ws://localhost/ws
```

**Protocol:** WebSocket (RFC 6455)

**Connection Flow:**
1. Client opens WebSocket connection
2. Server registers client in hub
3. Server sends welcome message
4. Client receives real-time events
5. Client can send ping for keepalive
6. Client closes connection when done

**Welcome Message (Server → Client):**

```json
{
  "type": "connection_established",
  "message": "Connected to AFS WebSocket",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### Event Types

**All events follow this structure:**

```json
{
  "type": "event_type_name",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {},
  "timestamp": "2024-01-15T10:30:05Z"
}
```

---

### Event: task_created

**Sent when:** User creates new task

**Payload:**

```json
{
  "type": "task_created",
  "task_id": 123,
  "payload": {
    "task_type": "query_optimization",
    "description": "Optimize monthly revenue report",
    "status": "pending"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### Event: agents_assigned

**Sent when:** Router selects agents for task

**Payload:**

```json
{
  "type": "agents_assigned",
  "task_id": 123,
  "payload": {
    "agents": ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"],
    "assignment_rationale": "High priority task with JOIN and aggregation. Assigned all agents for comprehensive analysis."
  },
  "timestamp": "2024-01-15T10:30:02Z"
}
```

---

### Event: fork_created

**Sent when:** Agent creates database fork (one event per agent)

**Payload:**

```json
{
  "type": "fork_created",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {
    "fork_id": "afs-fork-gemini-2.5-pro-task123-1699901234",
    "creation_time_seconds": 4.5
  },
  "timestamp": "2024-01-15T10:30:07Z"
}
```

---

### Event: analysis_completed

**Sent when:** Agent completes query analysis (one event per agent)

**Payload:**

```json
{
  "type": "analysis_completed",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {
    "insights_preview": [
      "Sequential scan on orders table processing 10k rows",
      "No index on (user_id, status) causing inefficient JOIN"
    ],
    "confidence": 0.92,
    "duration_seconds": 28
  },
  "timestamp": "2024-01-15T10:30:35Z"
}
```

---

### Event: proposal_submitted

**Sent when:** Agent generates optimization proposal (one event per agent)

**Payload:**

```json
{
  "type": "proposal_submitted",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {
    "proposal_id": 45,
    "proposal_type": "partial_index",
    "estimated_improvement": 80.0,
    "storage_overhead_mb": 12.0
  },
  "timestamp": "2024-01-15T10:31:30Z"
}
```

---

### Event: benchmark_started

**Sent when:** Agent begins benchmarking

**Payload:**

```json
{
  "type": "benchmark_started",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {
    "proposal_id": 45,
    "test_count": 4
  },
  "timestamp": "2024-01-15T10:31:35Z"
}
```

---

### Event: benchmark_completed

**Sent when:** Agent completes benchmarking (one event per agent)

**Payload:**

```json
{
  "type": "benchmark_completed",
  "task_id": 123,
  "agent_type": "gemini-2.5-pro",
  "payload": {
    "proposal_id": 45,
    "avg_improvement_pct": 82.6,
    "storage_overhead_mb": 12.0,
    "duration_seconds": 95
  },
  "timestamp": "2024-01-15T10:33:10Z"
}
```

---

### Event: consensus_reached

**Sent when:** Consensus engine selects winner

**Payload:**

```json
{
  "type": "consensus_reached",
  "task_id": 123,
  "payload": {
    "winning_proposal_id": 45,
    "winner_agent": "gemini-2.5-pro",
    "winning_score": 93.0,
    "rankings": {
      "gemini-2.5-pro": 1,
      "gemini-2.5-flash": 2,
      "gemini-2.0-flash": 3
    }
  },
  "timestamp": "2024-01-15T10:33:45Z"
}
```

---

### Event: optimization_applied

**Sent when:** Winning optimization applied to main DB

**Payload:**

```json
{
  "type": "optimization_applied",
  "task_id": 123,
  "payload": {
    "proposal_id": 45,
    "agent_type": "gemini-2.5-pro",
    "sql_executed": [
      "CREATE INDEX idx_orders_user_completed ON orders(user_id, status) WHERE status = 'completed'"
    ],
    "validation_passed": true,
    "actual_improvement_pct": 79.6
  },
  "timestamp": "2024-01-15T10:34:05Z"
}
```

---

### Event: task_completed

**Sent when:** Task fully completed (forks cleaned up)

**Payload:**

```json
{
  "type": "task_completed",
  "task_id": 123,
  "payload": {
    "status": "completed",
    "total_duration_seconds": 245,
    "final_improvement_pct": 79.6,
    "winning_agent": "gemini-2.5-pro"
  },
  "timestamp": "2024-01-15T10:34:15Z"
}
```

---

### Event: task_failed

**Sent when:** Task fails

**Payload:**

```json
{
  "type": "task_failed",
  "task_id": 123,
  "payload": {
    "error_code": "ALL_AGENTS_FAILED",
    "error_message": "All agents encountered errors during execution",
    "failed_agents": ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"],
    "partial_results": false
  },
  "timestamp": "2024-01-15T10:32:00Z"
}
```

---

### Client → Server Messages (Optional)

**Ping (Keepalive):**

```json
{
  "type": "ping"
}
```

**Server Response:**

```json
{
  "type": "pong",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Subscribe to Specific Task:**

```json
{
  "type": "subscribe",
  "task_id": 123
}
```

**Server Response:**

```json
{
  "type": "subscribed",
  "task_id": 123,
  "message": "You will receive updates for task 123"
}
```

---

## ❌ Error Responses

### Standard Error Format

**All error responses follow this structure:**

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {},
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

### Error Codes

**Validation Errors (400):**

**VALIDATION_ERROR:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "field_name": "Specific validation error message"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**INVALID_QUERY_SYNTAX:**
```json
{
  "error": {
    "code": "INVALID_QUERY_SYNTAX",
    "message": "Target query contains syntax errors",
    "details": {
      "query": "SELECT * FORM users",
      "syntax_error": "Syntax error near 'FORM'"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

**Not Found Errors (404):**

**TASK_NOT_FOUND:**
```json
{
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "Task with id 999 not found",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**PROPOSAL_NOT_FOUND:**
```json
{
  "error": {
    "code": "PROPOSAL_NOT_FOUND",
    "message": "Proposal with id 999 not found",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

**Conflict Errors (409):**

**DUPLICATE_TASK:**
```json
{
  "error": {
    "code": "DUPLICATE_TASK",
    "message": "Identical task submitted within last hour",
    "details": {
      "existing_task_id": 122,
      "cooldown_expires_at": "2024-01-15T11:25:00Z"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

**Server Errors (500):**

**INTERNAL_ERROR:**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "details": {
      "error_id": "uuid-1234-5678",
      "support_message": "Please contact support with this error ID"
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**DATABASE_ERROR:**
```json
{
  "error": {
    "code": "DATABASE_ERROR",
    "message": "Database connection failed",
    "details": {
      "retry_after_seconds": 30
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

## 📦 Data Transfer Objects (DTOs)

### Request DTOs

**CreateTaskRequest:**

```json
{
  "type": "string (required)",
  "description": "string (optional, max 500)",
  "target_query": "string (required, min 10)",
  "metadata": {
    "priority": "string (optional, default: medium)",
    "complexity_hint": "string (optional)",
    "target_tables": ["array of strings (optional)"],
    "user_preferences": {
      "max_storage_mb": "number (optional)",
      "risk_tolerance": "string (optional)"
    },
    "scoring_weights": {
      "performance": "number (optional, 0-1)",
      "storage": "number (optional, 0-1)",
      "complexity": "number (optional, 0-1)",
      "risk": "number (optional, 0-1)"
    }
  }
}
```

---

### Response DTOs

**TaskResponse:**

```json
{
  "id": "integer",
  "type": "string",
  "description": "string",
  "target_query": "string",
  "status": "string",
  "created_at": "ISO 8601 timestamp",
  "completed_at": "ISO 8601 timestamp or null",
  "metadata": "object",
  "links": {
    "self": "string (URL)",
    "agents": "string (URL)",
    "proposals": "string (URL)",
    "consensus": "string (URL)"
  }
}
```

**ProposalResponse:**

```json
{
  "id": "integer",
  "agent_type": "string",
  "agent_execution_id": "integer",
  "proposal_type": "string",
  "sql_commands": ["array of strings"],
  "rationale": "string",
  "estimated_impact": {
    "query_time_improvement": "number",
    "storage_overhead_mb": "number",
    "complexity": "string",
    "risk": "string"
  },
  "benchmark_summary": {
    "avg_improvement_pct": "number",
    "min_improvement_pct": "number",
    "max_improvement_pct": "number",
    "storage_overhead_mb": "number",
    "test_count": "integer"
  },
  "created_at": "ISO 8601 timestamp"
}
```

**BenchmarkResponse:**

```json
{
  "id": "integer",
  "query_name": "string",
  "query_executed": "string",
  "execution_time_ms": "number (2 decimals)",
  "rows_returned": "integer",
  "explain_plan": "object",
  "storage_impact_mb": "number or null",
  "improvement_pct": "number or null",
  "created_at": "ISO 8601 timestamp"
}
```

**ConsensusResponse:**

```json
{
  "task_id": "integer",
  "winning_proposal_id": "integer or null",
  "winner_agent": "string or null",
  "applied_to_main": "boolean",
  "scores": {
    "[agent_type]": {
      "proposal_id": "integer",
      "performance_score": "number (0-100)",
      "storage_score": "number (0-100)",
      "complexity_score": "number (0-100)",
      "risk_score": "number (0-100)",
      "weighted_total": "number (0-100)",
      "rank": "integer",
      "improvement_pct": "number",
      "storage_overhead_mb": "number"
    }
  },
  "decision_rationale": "string",
  "created_at": "ISO 8601 timestamp"
}
```

---

## 🔍 Additional Endpoints

### GET /health

**Purpose:** Health check for monitoring

**Response (200 OK):**

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "dependencies": {
    "database": "connected",
    "redis": "connected",
    "mcp_server": "available"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Response (503 Service Unavailable):**

```json
{
  "status": "unhealthy",
  "dependencies": {
    "database": "disconnected",
    "redis": "connected",
    "mcp_server": "unavailable"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### GET /stats

**Purpose:** System-wide statistics (optional, for demo)

**Response (200 OK):**

```json
{
  "tasks": {
    "total": 125,
    "completed": 98,
    "failed": 12,
    "in_progress": 15,
    "avg_duration_seconds": 210
  },
  "agents": {
    "gemini-2.5-pro": {
      "tasks_completed": 45,
      "success_rate": 0.96,
      "avg_improvement_pct": 78.5,
      "wins": 30
    },
    "gemini-2.0-flash": {
      "tasks_completed": 38,
      "success_rate": 0.89,
      "avg_improvement_pct": 65.2,
      "wins": 8
    },
    "gemini-2.5-flash": {
      "tasks_completed": 42,
      "success_rate": 0.93,
      "avg_improvement_pct": 85.3,
      "wins": 12
    }
  },
  "consensus": {
    "avg_score_spread": 15.2,
    "unanimous_decisions": 12,
    "close_decisions": 35
  }
}
```

---

## 🎯 Summary

This API specification provides:

**RESTful Endpoints:**
- Tasks: Create, retrieve, list (3 endpoints)
- Agents: List, status (2 endpoints)
- Optimizations: Proposals, benchmarks (2 endpoints)
- Consensus: Decision details (1 endpoint)
- System: Health, stats (2 endpoints)
- **Total: 10 REST endpoints**

**WebSocket Events:**
- 11 event types for real-time updates
- Bidirectional communication (ping/pong, subscribe)
- Task-specific event filtering

**Error Handling:**
- Standardized error format
- 8 distinct error codes
- Detailed validation messages
- Appropriate HTTP status codes

**Data Formats:**
- Consistent JSON structure
- ISO 8601 timestamps
- Numeric precision (2 decimals for times, scores)
- HATEOAS links for discoverability

**API Features:**
- Pagination (limit/offset)
- Filtering (status, type)
- Sorting (configurable)
- Hypermedia links (self, related resources)

---

**Related Documentation:**
- Previous: [05-CONSENSUS-BENCHMARKING.md](05-CONSENSUS-BENCHMARKING.md) 
  - Score calculations shown in responses
- See also: [02-DATA-MODEL.md](02-DATA-MODEL.md) 
  - Entity structures match DTOs
- See also: [01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md) 
  - Events match workflow steps

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```

---

## ✅ Documento `08-API-SPECIFICATION.md` Creado

**Contenido incluido:**
- ✅ **10 REST endpoints** completos (Tasks, Agents, Optimizations, Consensus, System)
- ✅ **11 WebSocket events** con payloads detallados
- ✅ **Request/Response schemas** con ejemplos JSON completos
- ✅ **Error handling** (8 códigos de error con ejemplos)
- ✅ **DTOs** (Data Transfer Objects) estructurados
- ✅ **Pagination, filtering, sorting** especificados
- ✅ **HATEOAS links** (hypermedia)
- ✅ **Authentication** (placeholder para futuro)
- ✅ **Sin código** (solo especificación con ejemplos JSON)

**Características destacadas:**
- Ejemplos JSON realistas basados en el flujo de Business Logic
- WebSocket events mapeados al workflow completo
- Error responses consistentes
- Health check endpoint para monitoring
- Stats endpoint para analytics

**¿Qué documento quieres escribir ahora?**
- 06-TIGER-CLOUD-MCP.md
- 07-LLM-INTEGRATION.md
- 09-FRONTEND-COMPONENTS.md
- 10-DEVELOPMENT-WORKFLOW.md
- 11-DEPLOYMENT-STRATEGY.md

🚀