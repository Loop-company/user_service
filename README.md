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

## Kafka data flow

User Service читает `user.events`:

- `user.registered` от Auth Service. Payload содержит `email`; сервис создает профиль с `ID = user_id`, дефолтным avatar и сгенерированным discriminator.

User Service публикует в `user.events`:

- `user.profile_updated` при изменении профиля;
- `user.settings_updated` при изменении настроек.

## Запуск

Docker Compose для локального запуска вынесен в репозиторий `loop_infra`.

Из корня `loop_infra`:

```bash
docker compose up --build
```

В этом репозитории остается код сервиса, `Dockerfile` и пример переменных окружения `user-service/.env.example`.
