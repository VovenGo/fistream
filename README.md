# fistream

`fistream` — приватные временные комнаты видеоконференцсвязи для русскоязычных пользователей.

После ввода `Имя` и `Пароль от сервиса` пользователь сразу попадает в комнату с функционалом Jitsi (аудио, камера, демонстрация экрана), без регистрации.

## Что реализовано

- Главная страница в стиле `fimstrimWWW` (только нужные блоки):
  - `Создать комнату`: `Имя`, `Пароль от сервиса`, кнопка `Войти`
  - `Войти в комнату`: `Имя`, `Код комнаты`, `Пароль от сервиса`, кнопка `Войти`
- Поле `Пароль комнаты` отсутствует.
- Поля `Имя`/`Код комнаты`/`Пароль от сервиса` отключают автоподстановку (`autocomplete=off`).
- API:
  - `POST /api/v1/rooms/create`
  - `POST /api/v1/rooms/join`
  - `POST /api/v1/rooms/{code}/close`
  - `GET /api/v1/config`
  - `GET /healthz`, `GET /readyz`
- Комнаты временные:
  - 6-символьный код `A-Z0-9`
  - авто-закрытие по TTL неактивности
  - ручное закрытие хостом
- Jitsi JWT выдается backend-ом после проверки `SERVICE_ACCESS_PASSWORD`.

## Быстрый старт (локально)

```bash
cp .env.example .env
docker compose up --build
```

Сервисы:
- app: [http://localhost:8080](http://localhost:8080)
- jitsi-web: [http://localhost:8000](http://localhost:8000)

## Переменные окружения

Основные:
- `SERVICE_ACCESS_PASSWORD`
- `DATABASE_URL`
- `ROOM_TTL`
- `API_TOKEN_SECRET`
- `JITSI_DOMAIN`
- `JITSI_APP_ID`
- `JITSI_APP_SECRET`
- `JITSI_AUDIENCE`
- `JITSI_SUBJECT`

## DevOps

- CI: `.github/workflows/ci.yml`
  - `gofmt` check
  - `go test ./...`
  - `go build ./cmd/api`
  - docker build
- Публикация image в Docker Hub: `.github/workflows/publish-image.yml`
- Прод деплой в Kubernetes с manual approval: `.github/workflows/deploy-prod.yml`
  - job использует `environment: production` (approval настраивается в GitHub Environment)

## Kubernetes

Манифесты: `deploy/k8s`
- `namespace.yaml`
- `postgres.yaml`
- `jitsi.yaml` (web/prosody/jicofo/jvb + UDP LB для JVB)
- `api.yaml`
- `ingress.yaml`

Подробности: `deploy/k8s/README.md`

## Тесты

```bash
go test ./...
```

Покрывает:
- валидацию сервисного пароля
- генерацию room code
- TTL-логику закрытия
- Jitsi JWT claims
- интеграционный create/join/close flow на HTTP-слое

