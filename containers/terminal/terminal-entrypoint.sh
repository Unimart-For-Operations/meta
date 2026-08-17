#!/usr/bin/env sh
set -e

# Which deployment to exec into, and which container inside it carries the shell.
SANDBOX_NAME="${SANDBOX_NAME:-}"
SANDBOX_CONTAINER="${SANDBOX_CONTAINER:-nix}"
SHELL_CMD="${SHELL_CMD:-/bin/sh}"

if [ -n "${SANDBOX_NAME}" ]; then
  # kubectl exec works against a Deployment via the deploy/<name> ref. The
  # pod's ServiceAccount grants pods/exec on its own namespace.
  exec ttyd --writable --port 7681 \
    kubectl exec -it "deploy/${SANDBOX_NAME}" -c "${SANDBOX_CONTAINER}" -- "${SHELL_CMD}"
fi

exec ttyd --writable --port 7681 /bin/sh