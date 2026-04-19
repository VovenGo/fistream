# Kubernetes deploy notes

1. Create namespace and secrets:
   - `kubectl apply -f namespace.yaml`
   - `kubectl -n fistream create secret generic fistream-secrets ...`
2. Apply infra:
   - `kubectl apply -f postgres.yaml`
   - `kubectl apply -f migrations-configmap.yaml`
   - `kubectl apply -f migrate.yaml`
   - `kubectl apply -f jitsi.yaml`
3. Apply API + ingress:
   - replace `REPLACE_IMAGE` in `api.yaml`
   - `kubectl apply -f api.yaml`
   - `kubectl apply -f ingress.yaml`

Required secret keys:
- `postgres-user`, `postgres-password`, `database-url`
- `service-access-password`, `api-token-secret`
- `jitsi-domain`, `jitsi-app-id`, `jitsi-app-secret`, `jitsi-audience`, `jitsi-subject`
- `jicofo-component-secret`, `jicofo-auth-password`, `jvb-auth-password`

Hosts:
- `fistream.example.com` -> web app/API
- `meet.fistream.example.com` -> Jitsi web/external_api.js

