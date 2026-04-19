#!/usr/bin/env python3
"""Set GitHub Actions repository secrets via GitHub REST API.

Usage:
  python scripts/set_github_secrets.py \
    --owner VovenGo --repo fistream --token <github_pat> \
    --secret DOCKERHUB_USERNAME=vovengo \
    --secret DOCKERHUB_TOKEN=... \
    --secret KUBE_CONFIG_DATA=...

Or provide --secret-file with KEY=VALUE lines.
"""

from __future__ import annotations

import argparse
import base64
import json
import sys
from pathlib import Path
from typing import Dict, Iterable, Tuple
from urllib import request
from urllib.error import HTTPError

from nacl import encoding, public


def parse_secret_line(line: str) -> Tuple[str, str] | None:
    stripped = line.strip()
    if not stripped or stripped.startswith("#"):
        return None
    if "=" not in stripped:
        raise ValueError(f"invalid secret line: {line!r}")
    key, value = stripped.split("=", 1)
    key = key.strip()
    value = value.strip()
    if not key:
        raise ValueError(f"empty secret key in line: {line!r}")
    return key, value


def load_secrets(inline: Iterable[str], secret_file: str | None) -> Dict[str, str]:
    secrets: Dict[str, str] = {}
    for item in inline:
        parsed = parse_secret_line(item)
        if parsed is None:
            continue
        key, value = parsed
        secrets[key] = value

    if secret_file:
        path = Path(secret_file)
        for line in path.read_text(encoding="utf-8").splitlines():
            parsed = parse_secret_line(line)
            if parsed is None:
                continue
            key, value = parsed
            secrets[key] = value

    if not secrets:
        raise ValueError("no secrets provided")
    return secrets


def github_request(url: str, token: str, method: str = "GET", payload: dict | None = None) -> dict:
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {token}",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "fistream-secret-uploader",
    }
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = request.Request(url=url, method=method, headers=headers, data=data)
    try:
        with request.urlopen(req, timeout=30) as resp:
            raw = resp.read()
            if not raw:
                return {}
            return json.loads(raw.decode("utf-8"))
    except HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"github api error {exc.code} {exc.reason}: {body}") from exc


def encrypt_secret(public_key_b64: str, secret_value: str) -> str:
    public_key = public.PublicKey(public_key_b64.encode("utf-8"), encoding.Base64Encoder())
    sealed_box = public.SealedBox(public_key)
    encrypted = sealed_box.encrypt(secret_value.encode("utf-8"))
    return base64.b64encode(encrypted).decode("utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description="Upload GitHub Actions repository secrets")
    parser.add_argument("--owner", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--token", required=True)
    parser.add_argument("--secret", action="append", default=[], help="KEY=VALUE")
    parser.add_argument("--secret-file", help="Path to file with KEY=VALUE lines")
    args = parser.parse_args()

    secrets = load_secrets(args.secret, args.secret_file)
    base_url = f"https://api.github.com/repos/{args.owner}/{args.repo}/actions/secrets"

    key_response = github_request(f"{base_url}/public-key", token=args.token)
    key_id = key_response.get("key_id")
    key = key_response.get("key")
    if not key_id or not key:
        raise RuntimeError("failed to read repo public key")

    for secret_name, secret_value in secrets.items():
        encrypted_value = encrypt_secret(key, secret_value)
        payload = {"encrypted_value": encrypted_value, "key_id": key_id}
        github_request(f"{base_url}/{secret_name}", token=args.token, method="PUT", payload=payload)
        print(f"updated secret: {secret_name}")

    list_response = github_request(base_url, token=args.token)
    names = sorted(item.get("name", "") for item in list_response.get("secrets", []))
    print("available secrets:")
    for name in names:
        if name:
            print(f"- {name}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover
        print(f"error: {exc}", file=sys.stderr)
        raise
