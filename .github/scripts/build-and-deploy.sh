#!/bin/bash

set -euo pipefail

echo "Starting static files deployment"

npm ci
npm run build

echo "Syncing with S3 bucket ${S3_BUCKET}"
aws s3 sync ./static "s3://${S3_BUCKET}" \
    --delete \
    --no-progress \
    --only-show-errors

TAG=$(git rev-parse --short HEAD)
if [ -z "$TAG" ]; then
    echo "Error: failed to get git commit hash"
    exit 1
fi

echo "Building container image with tag: ${TAG}"
echo "$TAG" > version

if ! docker buildx build . \
    --tag "$ECR_REPO:$TAG" \
    --platform linux/arm64 \
    --push \
    --provenance=false; then
    echo "Error: Docker build failed"
    exit 1
fi

echo "Getting image digest from ECR"
SHA_256=$(aws ecr describe-images \
    --repository-name lambda/ipv6test \
    --image-ids imageTag="$TAG" \
    --query 'imageDetails[0].imageDigest' \
    --output text)

if [ -z "$SHA_256" ] || [ "$SHA_256" = "null" ]; then
    echo "Error: Failed to get image digest"
    exit 1
fi

echo "Updating Lambda function code"
aws lambda update-function-code \
    --function-name "${FUNCTION_NAME}" \
    --image-uri "${ECR_REPO}@${SHA_256}"

wait_for_status() {
    local description="$1"
    local max_attempts=30
    local attempt=0
    local wait_time=10

    while [ $attempt -lt $max_attempts ]; do
        local current_status
        current_status=$(aws lambda get-function-configuration \
            --function-name "${FUNCTION_NAME}" \
            --query "LastUpdateStatus" \
            --output text 2>/dev/null) || true

        case "$current_status" in
            "Successful")
                echo "$description successful"
                return 0
                ;;
            "Failed")
                echo "$description failed"
                return 1
                ;;
            *)
                echo "Waiting for $description (attempt $((attempt+1))/$max_attempts, status: ${current_status:-unknown})"
                sleep $wait_time
                ((attempt++))
                ;;
        esac
    done

    echo "Timeout waiting for $description after $((max_attempts * wait_time)) seconds"
    return 1
}

if ! wait_for_status 'Function code update'; then
    echo "Error: Function code update failed"
    exit 1
fi

echo "Publishing new version"
VERSION=$(aws lambda publish-version \
            --function-name "${FUNCTION_NAME}" \
            --description "New deployment" \
            --query 'Version' \
            --output text)

if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
    echo "Error: failed to publish version"
    exit 1
fi

if ! wait_for_status "Version ${VERSION}"; then
    echo "Error: Version publication failed"
    exit 1
fi

echo "Updating alias 'main' to version ${VERSION}"
aws lambda update-alias \
    --function-name "${FUNCTION_NAME}" \
    --function-version "${VERSION}" \
    --name main

echo 'Deployment completed successfully'
