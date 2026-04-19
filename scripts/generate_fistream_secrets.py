#!/usr/bin/env python3
"""Generate strong values for fistream Kubernetes secret and print KEY=VALUE lines."""

from __future__ import annotations

import argparse
import secrets
import string


def token(length: int = 48) -> str:
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(length))


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--postgres-user", default="fistream")
    parser.add_argument("--jitsi-app-id", default="fistream")
    parser.add_argument("--jitsi-audience", default="fistream")
    parser.add_argument("--jitsi-subject", default="meet.jitsi")
    parser.add_argument("--jitsi-domain", default="fistream.vovengo.com:8443")
    parser.add_argument("--allowed-origins", default="https://fistream.vovengo.com")
    args = parser.parse_args()

    postgres_password = token(32)
    values = {
        "postgres-user": args.postgres_user,
        "postgres-password": postgres_password,
        "database-url": f"postgres://{args.postgres_user}:{postgres_password}@postgres.fistream.svc.cluster.local:5432/fistream?sslmode=disable",
        "service-access-password": token(20),
        "api-token-secret": token(64),
        "jitsi-domain": args.jitsi_domain,
        "jitsi-app-id": args.jitsi_app_id,
        "jitsi-app-secret": token(64),
        "jitsi-audience": args.jitsi_audience,
        "jitsi-subject": args.jitsi_subject,
        "jicofo-component-secret": token(48),
        "jicofo-auth-password": token(32),
        "jvb-auth-password": token(32),
        "allowed-origins": args.allowed_origins,
    }

    for key, value in values.items():
        print(f"{key}={value}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
