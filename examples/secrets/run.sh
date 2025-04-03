#!/bin/bash

# List of required secrets
REQUIRED_SECRETS=(
  "MY_SECRET"
  "MY_OTHER_SECRET"
  "HELLO_WORLD"
  "NOT_SECRET"
)

missing_secrets=()
defined_secrets=()
error_value_secrets=()

echo "Checking for required secrets..."

for secret in "${REQUIRED_SECRETS[@]}"; do
  if [ -z "${!secret}" ]; then
    missing_secrets+=("$secret")
    echo "❌ Missing secret: $secret"
  else
    defined_secrets+=("$secret")
    # Check if the secret value contains "error"
    if [ "${!secret}" = "error" ]; then
      error_value_secrets+=("$secret")
      echo "❌ Secret $secret contains invalid value 'error'"
    else
      echo "✅ Found secret: $secret = ${!secret}"
    fi
  fi
done

if [ ${#missing_secrets[@]} -ne 0 ]; then
  echo "Error: Missing ${#missing_secrets[@]} required secrets"
  exit 1
fi
if [ ${#error_value_secrets[@]} -ne 0 ]; then
  echo "Error: ${#error_value_secrets[@]} secrets contain the invalid value 'error'"
  exit 1
fi
echo "Success: All required secrets are available and valid"
