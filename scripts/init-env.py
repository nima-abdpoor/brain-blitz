#!/usr/bin/env python3
"""Create machine-local secrets without overwriting an existing deployment."""
from pathlib import Path
import os
import secrets

root = Path(__file__).resolve().parents[1]
path = root / ".env"
try:
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
except FileExistsError:
    print(".env already exists; keeping its secrets and settings.")
else:
    with os.fdopen(fd, "w", encoding="utf-8", newline="\n") as env:
        env.write(
            f"POSTGRES_PASSWORD={secrets.token_hex(32)}\n"
            f"JWT_SECRET={secrets.token_hex(32)}\n"
            "HTTP_BIND=127.0.0.1\nHTTP_PORT=8088\n"
            "GOPROXY=https://proxy.golang.org|https://goproxy.cn|direct\n"
        )
    print("Created .env. Keep it private and back it up with your database backups.")
