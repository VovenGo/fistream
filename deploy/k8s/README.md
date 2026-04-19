# Kubernetes deploy notes (k3s + Caddy edge)

This project runs on k3s with `traefik` disabled.  
Traffic is terminated by host-level Caddy on:

- `https://fistream.vovengo.com` -> app/API (`NodePort 30080`)
- `https://fistream.vovengo.com:8443` -> Jitsi web (`NodePort 30443`)

JVB media uses UDP `31000` on the node.

## 1) Create namespace and secrets

```bash
kubectl apply -f namespace.yaml
kubectl -n fistream create secret generic fistream-secrets \
  --from-literal=postgres-user='fistream' \
  --from-literal=postgres-password='<strong-password>' \
  --from-literal=database-url='postgres://fistream:<strong-password>@postgres.fistream.svc.cluster.local:5432/fistream?sslmode=disable' \
  --from-literal=service-access-password='<service-password>' \
  --from-literal=api-token-secret='<api-token-secret>' \
  --from-literal=jitsi-domain='fistream.vovengo.com:8443' \
  --from-literal=jitsi-app-id='fistream' \
  --from-literal=jitsi-app-secret='<jitsi-app-secret>' \
  --from-literal=jitsi-audience='fistream' \
  --from-literal=jitsi-subject='meet.jitsi' \
  --from-literal=jicofo-component-secret='<jicofo-component-secret>' \
  --from-literal=jicofo-auth-password='<jicofo-auth-password>' \
  --from-literal=jvb-auth-password='<jvb-auth-password>' \
  --from-literal=allowed-origins='https://fistream.vovengo.com'
```

## 2) Apply infra and app

```bash
kubectl apply -f postgres.yaml
kubectl apply -f migrations-configmap.yaml
kubectl delete job -n fistream fistream-migrate --ignore-not-found=true
kubectl apply -f migrate.yaml
kubectl wait --for=condition=complete --timeout=300s job/fistream-migrate -n fistream
kubectl apply -f jitsi.yaml
kubectl apply -f api.yaml
```

If using CI/CD, `api.yaml` is rendered with image tag from GitHub Actions.

## 3) Required secret keys

- `postgres-user`, `postgres-password`, `database-url`
- `service-access-password`, `api-token-secret`
- `jitsi-domain`, `jitsi-app-id`, `jitsi-app-secret`, `jitsi-audience`, `jitsi-subject`
- `jicofo-component-secret`, `jicofo-auth-password`, `jvb-auth-password`
- `allowed-origins`

## 4) NodePorts expected by Caddy

- API service: `30080/tcp`
- Jitsi web: `30443/tcp`
- JVB media: `31000/udp`
