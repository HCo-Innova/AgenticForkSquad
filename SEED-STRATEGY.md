# 🚀 Seed Strategy - Professional Implementation

## Overview
This project uses a **two-layer database initialization** pattern:

1. **Goose Migrations** (`backend/migrations/*.sql`)
   - Tracks execution via `goose_db_version` table
   - Automatically skips already-applied migrations
   - Idempotent and safe to re-run

2. **Seed Data** (`backend/cmd/api/seed/main.go`)
   - Tracks execution via `seed_metadata` table
   - Automatically skips if already seeded (like Goose)
   - Idempotent TRUNCATE + INSERT pattern

---

## Execution Flow

### Docker Compose Startup
```
docker compose up

↓

migrate service runs:
  - Executes 001_create_schema.sql
  - Executes 002_afs_tables.sql
  - Executes 003_seed_metadata.sql
  - Creates goose_db_version table with version tracking
  - Result: 3 rows in goose_db_version (all marked applied=true)

↓ (after migrate completes successfully)

seed service runs:
  - Connects to database
  - Checks seed_metadata table for version "v1"
  - If NOT found → inserts 1000 users + 10000 orders/payments
  - If FOUND → skips with message "already applied"
  - Records execution in seed_metadata for future runs

↓ (after seed completes)

backend service starts:
  - Depends on seed.condition=service_completed_successfully
  - Finds 1000 users ready to use
  - Mounts tiger-config volume for credentials persistence
```

---

## Migration Tracking Tables

### `goose_db_version` (Auto-Created by Goose)
```sql
id | version_id | is_applied | tstamp
---|------------|-----------|-----------
1  | 1          | true      | 2025-11-06 10:00:00
2  | 2          | true      | 2025-11-06 10:00:05
3  | 3          | true      | 2025-11-06 10:00:10
```

### `seed_metadata` (Created by Migration 003)
```sql
id | seed_version | seeded_at          | row_count_users | row_count_orders | status
---|--------------|--------------------|-----------------|-----------------|---------
1  | v1           | 2025-11-06 10:00:15| 1000            | 10000            | success
```

---

## Idempotency Pattern

### First Run
```
$ docker compose up
[migrate] Executing: 001_create_schema.sql ✓
[migrate] Executing: 002_afs_tables.sql ✓
[migrate] Executing: 003_seed_metadata.sql ✓
[seed]    Seeding v1... [1000 users, 10000 orders created] ✓
[backend] Started with healthy data
```

### Second Run (Same Environment)
```
$ docker compose up
[migrate] goose status:
           001: APPLIED (skipped)
           002: APPLIED (skipped)
           003: APPLIED (skipped)
[seed]    ⏭️  Seed already applied (version v1). Skipping.
[backend] Started (no re-seeding)
```

### Update Scenario (New Seed Version)
When you want to re-seed with different data:
```
# In seed/main.go
const seedVersion = "v2"  // Bump version
```

Next run:
```
[seed] Checking seed_metadata for v2... not found
[seed] TRUNCATE + INSERT 1000 users, 10000 orders
[seed] Record v2 in seed_metadata ✓
```

---

## Manual Commands

### Check Migration Status
```bash
docker compose exec backend goose -dir ./migrations postgres "$DATABASE_URL" status
```

### Check Seed Status
```bash
docker compose exec backend tiger service query wuj5xa6zpz "SELECT * FROM seed_metadata ORDER BY seeded_at DESC;"
```

### Force Re-seed (Reset All Data)
```bash
# Delete seed metadata entry
docker compose exec backend tiger service query wuj5xa6zpz "DELETE FROM seed_metadata WHERE seed_version = 'v1';"

# Next docker compose up will re-seed automatically
docker compose up seed
```

### Manual Seed Without Docker
```bash
cd backend
DATABASE_URL="postgres://..." go run cmd/api/seed/main.go
```

---

## Why This Approach?

✅ **Safe**: Won't re-seed on every container restart
✅ **Fast**: Skips if already applied
✅ **Professional**: Uses standard tracking pattern (like Goose)
✅ **Debuggable**: Queryable metadata tables
✅ **Scalable**: Easy to add seed v2, v3, etc
✅ **CI/CD Ready**: Works in automated pipelines
