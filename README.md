# User Service

User profile service for profile data, settings, and online status.

## Project layout

- `user-service/` - Go application source code
- `user-service/.env.example` - example local environment variables
- `.github/workflows/ci.yml` - CI/CD pipeline for the lab
- `docker-compose.yml` - local infrastructure for the service

## Local run

1. Copy `user-service/.env.example` to `user-service/.env` and set your local values.
2. Build the containers:

```powershell
docker compose build
```

3. Start the stack:

```powershell
docker compose up -d
```

4. Check containers:

```powershell
docker compose ps
```

5. Stop the stack:

```powershell
docker compose down
```

## Example environment variables

```env
PORT=8081
POSTGRES_HOST=db-user
POSTGRES_PORT=5432
POSTGRES_DB=user_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=change-me
REDIS_HOST=redis-user
REDIS_PORT=6379
REDIS_PASSWORD=
KAFKA_BROKERS=kafka:9092
```

## CI/CD for the lab

The GitHub Actions pipeline contains the required jobs:

- `build`
- `lint`
- `test`
- `docker_build`
- `docker_push`

The `test` job uploads coverage artifacts and fails if coverage is below `50%`.
The `docker_push` job pushes the image only on `push` events and uses GitHub secrets for Docker Hub authentication.

## GitHub secrets and variables

Add these repository secrets before pushing Docker images:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

Optional repository variable:

- `DOCKERHUB_NAMESPACE`
