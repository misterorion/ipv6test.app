#!/bin/bash

set -euo pipefail

TAG=$(git rev-parse --short HEAD)

if [ -z "$TAG" ]; then
    echo "Error: failed to get git commit hash"
    exit 1
fi

echo "Building container image with tag: ${TAG}"

echo "$TAG" > version
echo "tag=$TAG" >> "$GITHUB_OUTPUT"

if ! docker buildx build \
    --tag "${ECR_REPO}:${TAG}" \
    --platform linux/arm64 \
    --provenance=false \
    --push .; then
    echo "Error: Docker build failed" >&2
    exit 1
fi
