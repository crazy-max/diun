#!/bin/sh

# Exit immediately if a command fails
set -o errexit

# Treat unset variables as an error
set -o nounset

# If CA_CERT env var is set and points to an existing file
if [ -n "${CA_CERTIFICATE:-}" ] && [ -f "${CA_CERTIFICATE:-}" ]; then
    echo "Installing custom CA certificate from $CA_CERTIFICATE"
    cp "$CA_CERTIFICATE" /usr/local/share/ca-certificates/custom-ca.crt
    update-ca-certificates
fi

# Execute the command passed as arguments
exec diun "$@"