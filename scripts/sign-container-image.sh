#!/usr/bin/env bash

set -euo pipefail

image_ref=${1:?container image digest reference is required}
workflow_ref=${GITHUB_WORKFLOW_REF:?GitHub workflow reference is required}
certificate_identity="https://github.com/$workflow_ref"
certificate_issuer="https://token.actions.githubusercontent.com"

cosign sign --yes "$image_ref"
cosign verify \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "$certificate_issuer" \
  "$image_ref"
