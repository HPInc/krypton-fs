#!/bin/sh
LINTER_IMAGE=golangci/golangci-lint:v2.1.6-alpine
DIR=${1:-$(pwd)}

# flag overrides
# excluding S1034. See https://staticcheck.io/docs/checks#S1034
docker run --rm \
  -v "$DIR":/service \
  -w /service \
  "$LINTER_IMAGE" golangci-lint run \
  -v --tests=false --timeout=5m
