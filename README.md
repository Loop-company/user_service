# User Service

gRPC service for user profiles, statuses, and settings.

## Architecture Role

- Receives user API calls from HTTP Gateway over gRPC.
- Consumes `user.registered` events from Kafka and creates user profiles.
- Publishes profile and settings changes to Kafka for Analytics Service.
- Stores profile data in PostgreSQL and presence/status data in Redis.

## Kafka Events

User Service publishes the shared analytics event envelope to `user.events`:

```json
{
  "event_id": "uuid",
  "user_id": "user-guid",
  "event_type": "user.profile_updated",
  "source_service": "user-service",
  "payload": {
    "changes": {
      "name": "new-name"
    }
  },
  "occurred_at": "2026-05-07T12:00:00Z"
}
```

## Configuration

```env
GRPC_PORT=50052
POSTGRES_HOST=db-user
POSTGRES_PORT=5432
POSTGRES_DB=userdb
POSTGRES_USER=user
POSTGRES_PASSWORD=user_password
REDIS_HOST=redis-user
REDIS_PORT=6379
KAFKA_BROKERS=kafka:9092
```

## CI

The GitHub Actions workflow runs build, golangci-lint, tests with coverage, Docker build, and Docker push.
