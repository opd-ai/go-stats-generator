# PostgreSQL Backend Configuration

The `go-stats-generator` API server supports PostgreSQL as a persistent storage backend for analysis results.

## Configuration

### Option 1: Configuration File

Create or update `.go-stats-generator.yaml`:

```yaml
storage:
  type: postgres
  postgres_connection_string: "postgres://username:password@hostname:5432/dbname?sslmode=disable"
```

### Option 2: Environment Variables

```bash
export POSTGRES_CONNECTION_STRING="postgres://username:password@hostname:5432/dbname?sslmode=disable"
```

### Option 3: Programmatic Configuration

```go
import (
    "github.com/opd-ai/go-stats-generator/internal/api/storage"
    "github.com/opd-ai/go-stats-generator/internal/config"
)

cfg := config.DefaultConfig()
cfg.Storage.Type = "postgres"
cfg.Storage.PostgresConnectionString = "postgres://user:pass@localhost:5432/dbname?sslmode=disable"

store := storage.New(cfg)
defer func() {
    if pg, ok := store.(*storage.Postgres); ok {
        pg.Close()
    }
}()
```

## Connection String Format

The connection string follows the PostgreSQL standard format:

```
postgres://username:password@hostname:port/database?param=value
```

### Common Parameters

- `sslmode`: Connection encryption mode (`disable`, `require`, `verify-ca`, `verify-full`)
- `connect_timeout`: Connection timeout in seconds
- `application_name`: Application identifier for PostgreSQL logs

### Example Connection Strings

```bash
# Local development (no SSL)
postgres://postgres:postgres@localhost:5432/go_stats_dev?sslmode=disable

# Production with SSL
postgres://app_user:secure_password@prod-db.example.com:5432/go_stats_prod?sslmode=require

# Connection timeout
postgres://user:pass@db:5432/dbname?sslmode=disable&connect_timeout=10
```

## Database Setup

### Manual Setup

```sql
-- Create database
CREATE DATABASE go_stats_prod;

-- Create user
CREATE USER go_stats_user WITH PASSWORD 'secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE go_stats_prod TO go_stats_user;
```

### Automatic Schema Creation

The PostgreSQL backend automatically creates the required table on first connection:

```sql
CREATE TABLE IF NOT EXISTS analysis_results (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    report JSONB,
    error TEXT
);
```

## Usage with API Server

Start the API server with PostgreSQL backend:

```bash
# Using environment variable
export POSTGRES_CONNECTION_STRING="postgres://user:pass@localhost:5432/go_stats?sslmode=disable"
go-stats-generator serve --enable-api

# Or with config file
go-stats-generator serve --enable-api --config .go-stats-generator.yaml
```

## Fallback Behavior

If PostgreSQL connection fails, the storage factory automatically falls back to in-memory storage. This ensures the API server remains operational even when the database is unavailable.

## Testing

Run tests with PostgreSQL backend:

```bash
# Set test database connection
export POSTGRES_TEST_URL="postgres://postgres:postgres@localhost:5432/go_stats_test?sslmode=disable"

# Run storage tests
go test ./internal/api/storage/... -v

# Tests will skip gracefully if PostgreSQL is unavailable
```

## Performance Considerations

- **Connection Pooling**: The backend uses Go's `database/sql` connection pooling automatically
- **JSONB Storage**: Reports are stored as JSONB for efficient querying and indexing
- **Thread Safety**: All operations are protected with read/write mutexes

## Production Recommendations

1. **Use SSL/TLS**: Always use `sslmode=require` or higher in production
2. **Connection Limits**: Configure PostgreSQL `max_connections` based on your load
3. **Monitoring**: Monitor database performance and query times
4. **Backups**: Implement regular database backups for disaster recovery
5. **Credentials**: Store connection strings in secure vaults (e.g., HashiCorp Vault, AWS Secrets Manager)

## Troubleshooting

### Connection Refused

```
failed to ping database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solution**: Verify PostgreSQL is running and accessible:
```bash
psql -h localhost -U postgres -d postgres
```

### Authentication Failed

```
pq: password authentication failed for user "username"
```

**Solution**: Verify credentials and user permissions in PostgreSQL.

### Database Does Not Exist

```
pq: database "dbname" does not exist
```

**Solution**: Create the database manually using `CREATE DATABASE dbname;`

## Migration from In-Memory

To migrate from in-memory to PostgreSQL backend:

1. Update configuration to use PostgreSQL
2. Restart the API server
3. Existing in-memory data is not migrated automatically
4. Consider implementing a data export/import workflow if persistence is required
