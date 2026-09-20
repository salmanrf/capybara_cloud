#!/usr/bin/env sh
set -e
case "$1" in
  dev)  ENV=development ;;
  prod) ENV=production ;;
  *) echo "usage: $0 (dev|prod)" >&2; exit 1 ;;
esac
make "$1"
LOG_FILE=apps.backend.log ENV=$ENV ./bin/backend
