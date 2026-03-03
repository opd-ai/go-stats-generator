# MongoDB Backend Configuration

The `go-stats-generator` API server supports MongoDB as a persistent storage backend for analysis results.

## Configuration

### Option 1: Configuration File

Create or update `.go-stats-generator.yaml`:

```yaml
storage:
  type: mongo
  mongo_connection_string: "mongodb://username:password@hostname:27017/go_stats_generator?authSource=admin"
```

### Option 2: Environment Variables

```bash
export MONGO_CONNECTION_STRING="mongodb://username:password@hostname:27017/go_stats_generator?authSource=admin"
```

### Option 3: Programmatic Configuration

```go
import (
    "github.com/opd-ai/go-stats-generator/internal/api/storage"
    "github.com/opd-ai/go-stats-generator/internal/config"
)

cfg := config.DefaultConfig()
cfg.Storage.Type = "mongo"
cfg.Storage.MongoConnectionString = "mongodb://localhost:27017/go_stats_generator"

store := storage.New(cfg)
defer func() {
    if mg, ok := store.(*storage.Mongo); ok {
        mg.Close()
    }
}()
```

## Connection String Format

The connection string follows the MongoDB standard URI format:

```
mongodb://[username:password@]host[:port][/database][?options]
```

### Common Parameters

- `authSource`: Authentication database (e.g., `admin`)
- `replicaSet`: Replica set name for high availability
- `ssl`: Enable SSL/TLS encryption (`true` or `false`)
- `connectTimeoutMS`: Connection timeout in milliseconds
- `maxPoolSize`: Maximum connection pool size

### Example Connection Strings

```bash
# Local development (no authentication)
mongodb://localhost:27017/go_stats_dev

# With authentication
mongodb://app_user:secure_password@localhost:27017/go_stats_prod?authSource=admin

# MongoDB Atlas (cloud)
mongodb+srv://username:password@cluster0.mongodb.net/go_stats_prod?retryWrites=true&w=majority

# Replica set with SSL
mongodb://user:pass@host1:27017,host2:27017,host3:27017/go_stats_prod?replicaSet=rs0&ssl=true

# Connection timeout
mongodb://localhost:27017/go_stats_dev?connectTimeoutMS=5000
```

## Database Setup

### Manual Setup

```javascript
// Connect to MongoDB shell
mongosh

// Create database and user
use go_stats_generator

db.createUser({
  user: "go_stats_user",
  pwd: "secure_password",
  roles: [
    { role: "readWrite", db: "go_stats_generator" }
  ]
})
```

### Automatic Schema Creation

The MongoDB backend automatically creates the required collection and indexes on first connection:

- **Database**: `go_stats_generator`
- **Collection**: `analysis_results`
- **Index**: Unique index on `_id` field (automatic)

Document structure:
```json
{
  "_id": "analysis-uuid",
  "status": "completed",
  "report": { /* BSON encoded Report object */ },
  "error": null
}
```

## Usage with API Server

Start the API server with MongoDB backend:

```bash
# Using environment variable
export MONGO_CONNECTION_STRING="mongodb://localhost:27017/go_stats_generator"
go-stats-generator serve --enable-api

# Or with config file
go-stats-generator serve --enable-api --config .go-stats-generator.yaml
```

## Fallback Behavior

If MongoDB connection fails, the storage factory automatically falls back to in-memory storage. This ensures the API server remains operational even when the database is unavailable.

## Testing

Run tests with MongoDB backend:

```bash
# Set test database connection
export MONGO_TEST_CONNECTION_STRING="mongodb://localhost:27017/go_stats_test"

# Run storage tests
go test ./internal/api/storage/... -v -run TestMongo

# Tests will skip gracefully if MongoDB is unavailable
```

## Performance Considerations

- **Connection Pooling**: The backend uses the official MongoDB driver's built-in connection pooling
- **BSON Storage**: Reports are stored as BSON for efficient querying and native MongoDB data types
- **Thread Safety**: All operations are protected with read/write mutexes
- **Context Timeouts**: All operations have configurable timeouts (5-10 seconds default)

## Production Recommendations

1. **Use Replica Sets**: Deploy MongoDB in replica set mode for high availability
2. **Enable Authentication**: Always use authentication in production environments
3. **SSL/TLS**: Enable SSL/TLS encryption for data in transit
4. **Connection Limits**: Configure `maxPoolSize` based on your concurrent load
5. **Monitoring**: Monitor database performance using MongoDB Atlas or Ops Manager
6. **Backups**: Implement regular database backups (use `mongodump` or MongoDB Atlas automated backups)
7. **Credentials**: Store connection strings in secure vaults (e.g., HashiCorp Vault, AWS Secrets Manager)

## MongoDB Atlas (Cloud)

For cloud deployment, MongoDB Atlas provides a fully managed solution:

```yaml
storage:
  type: mongo
  mongo_connection_string: "mongodb+srv://username:password@cluster0.mongodb.net/go_stats_generator?retryWrites=true&w=majority"
```

**Benefits**:
- Automated backups and point-in-time recovery
- Built-in monitoring and alerting
- Auto-scaling capabilities
- Global distribution with multi-region deployments

## Troubleshooting

### Connection Timeout

```
failed to connect to MongoDB: server selection error: context deadline exceeded
```

**Solution**: Verify MongoDB is running and accessible:
```bash
mongosh mongodb://localhost:27017
```

### Authentication Failed

```
failed to ping MongoDB: (Unauthorized) not authorized on admin to execute command
```

**Solution**: Verify credentials and user permissions in MongoDB.

### Database Connection Refused

```
failed to connect to MongoDB: dial tcp 127.0.0.1:27017: connect: connection refused
```

**Solution**: Ensure MongoDB service is running:
```bash
# Linux/macOS
sudo systemctl status mongod

# Or check process
ps aux | grep mongod
```

### SSL/TLS Certificate Errors

```
failed to connect to MongoDB: x509: certificate signed by unknown authority
```

**Solution**: Add `ssl=true&tlsInsecure=true` to connection string (development only) or provide proper CA certificates.

## Migration from In-Memory or PostgreSQL

To migrate from another backend to MongoDB:

1. Update configuration to use MongoDB backend
2. Restart the API server
3. Existing data is not migrated automatically
4. For data migration, export from old backend and import to MongoDB:
   - Export analysis results via API or direct database query
   - Import to MongoDB using `mongoimport` or programmatic insertion

## Advanced Configuration

### Custom Database Name

The default database name is `go_stats_generator`. To use a custom name, include it in the connection string:

```
mongodb://localhost:27017/my_custom_db
```

### Read/Write Preferences

For replica sets, configure read preference in the connection string:

```
mongodb://host1:27017,host2:27017/go_stats_generator?readPreference=primaryPreferred
```

### Connection Pool Settings

```
mongodb://localhost:27017/go_stats_generator?maxPoolSize=100&minPoolSize=10
```

## Comparison: MongoDB vs PostgreSQL vs In-Memory

| Feature | MongoDB | PostgreSQL | In-Memory |
|---------|---------|------------|-----------|
| Persistence | ✅ Yes | ✅ Yes | ❌ No |
| Schema Flexibility | ✅ High | ⚠️ Medium | ✅ High |
| Query Performance | ✅ Excellent | ✅ Excellent | ✅ Fastest |
| Horizontal Scaling | ✅ Native sharding | ⚠️ Requires setup | ❌ N/A |
| ACID Transactions | ✅ Yes (4.0+) | ✅ Yes | ❌ No |
| Setup Complexity | ⚠️ Medium | ⚠️ Medium | ✅ Zero |
| Best For | Cloud-native, JSON-heavy | Relational data | Development/testing |
