# User Service

User Service хранит пользовательские профили, статусы и настройки. HTTP Gateway обращается к сервису по gRPC. Кроме gRPC, сервис читает событие регистрации из Kafka и публикует события об изменениях профиля/настроек для аналитики.

## Место в архитектуре

```text
Client -> HTTP Gateway -> User Service (gRPC)
                            |
                            +-> PostgreSQL: profiles and settings
                            +-> Redis: cache/status data
                            +-> Kafka consumer: user.events / user.registered
                            +-> Kafka producer: user.events / profile and settings events
```

User Service:

- принимает gRPC-вызовы от HTTP Gateway;
- создает профиль после события `user.registered` от Auth Service;
- хранит профиль и настройки пользователя в PostgreSQL;
- использует Redis для cache/status данных;
- публикует пользовательские события в Kafka для Analytics Service.

## gRPC API

Контракт описан в `user-service/proto/user.proto`.

- `GetProfile(user_id)` - возвращает профиль пользователя.
- `UpdateName(user_id, name)` - обновляет имя пользователя.
- `UpdateStatus(user_id, status)` - обновляет статус.
- `GetSettings(user_id)` - возвращает настройки пользователя.
- `UpdateSettings(user_id, settings)` - обновляет настройки.

HTTP Gateway вызывает эти методы для protected HTTP routes. Идентичность пользователя проверяется gateway через Auth Service, а в User Service передается `user_id` в gRPC request.

## Kafka data flow

User Service читает `user.events`:

- `user.registered` от Auth Service. Payload содержит `email`; сервис создает профиль с `ID = user_id`, дефолтным avatar и сгенерированным discriminator.

User Service публикует в `user.events`:

- `user.profile_updated` при изменении профиля;
- `user.settings_updated` при изменении настроек.

Envelope событий:

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

Analytics Service читает этот же топик и сохраняет события для отчетов.

## Переменные окружения

Пример есть в `user-service/.env.example`.

```env
GRPC_PORT=50052

POSTGRES_HOST=db-user
POSTGRES_PORT=5432
POSTGRES_DB=userdb
POSTGRES_USER=user
POSTGRES_PASSWORD=user_password

REDIS_HOST=redis-user
REDIS_PORT=6379
REDIS_PASSWORD=

KAFKA_BROKERS=kafka:9092
```

## Запуск

Для запуска User Service через Docker Compose из корня репозитория:

```bash
docker compose up --build
```

Этот compose поднимает User Service, PostgreSQL и Redis. Для полноценной работы Kafka consumer/producer Kafka должна быть доступна по адресу из `KAFKA_BROKERS`.

Для запуска всего backend-стека используется Docker Compose в репозитории HTTP Gateway:

```bash
docker compose up --build
```
