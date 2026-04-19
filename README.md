# fistream

`fistream` — приватные временные комнаты видеосвязи (аудио/видео/демонстрация экрана) с входом по `Имя + Пароль от сервиса` и встроенным Jitsi без регистрации.

## Что уже реализовано

- Главная страница в стиле `fimstrimWWW` с двумя сценариями:
  - `Создать комнату`: `Имя`, `Пароль от сервиса`, кнопка `Войти`
  - `Войти в комнату`: `Имя`, `Код комнаты`, `Пароль от сервиса`, кнопка `Войти`
- Поле `Пароль комнаты` отсутствует в UI и API.
- Поля ввода отключают автоподстановку (`autocomplete=off`).
- Backend на Go + PostgreSQL:
  - `POST /api/v1/rooms/create`
  - `POST /api/v1/rooms/join`
  - `POST /api/v1/rooms/{code}/close`
  - `GET /api/v1/config`, `GET /healthz`, `GET /readyz`
- Комнаты временные (`A-Z0-9`, 6 символов), с TTL и ручным закрытием хостом.
- Jitsi JWT выдается backend-ом после проверки `SERVICE_ACCESS_PASSWORD`.

## Локальный запуск

```bash
cp .env.example .env
docker compose up --build
```

Сервисы:
- app: [http://localhost:8080](http://localhost:8080)
- jitsi-web: [http://localhost:8000](http://localhost:8000)

## Production схема

- Публичный домен app/API: `https://fistream.vovengo.com`
- Публичный домен Jitsi (тот же host, отдельный порт): `https://fistream.vovengo.com:8443`
- Kubernetes: k3s (без Traefik), сервисы публикуются через NodePort.
- Edge TLS/Reverse proxy: Caddy на хосте.

Подробности: [deploy/k8s/README.md](deploy/k8s/README.md)

## GitHub Secrets через API

Скрипт:
- `scripts/set_github_secrets.py`

Зависимость:
```bash
python -m pip install -r scripts/requirements.txt
```

Пример:
```bash
python scripts/set_github_secrets.py \
  --owner VovenGo \
  --repo fistream \
  --token <GITHUB_PAT> \
  --secret DOCKERHUB_USERNAME=vovengo \
  --secret DOCKERHUB_TOKEN=<DOCKERHUB_TOKEN> \
  --secret KUBE_CONFIG_DATA=<BASE64_KUBECONFIG>
```

## Генерация `fistream-secrets`

Скрипт генерации:
```bash
python scripts/generate_fistream_secrets.py > .secrets.env
```

Далее создать secret:
```bash
kubectl -n fistream create secret generic fistream-secrets \
  --from-env-file=.secrets.env
```

## CI/CD

- `.github/workflows/ci.yml` — tests/build
- `.github/workflows/publish-image.yml` — push образа в Docker Hub
- `.github/workflows/deploy-prod.yml` — deploy в k3s с `KUBE_CONFIG_DATA`

## Тесты

```bash
go test ./...
```
