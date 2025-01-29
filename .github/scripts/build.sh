#!/bin/bash

set -e

# 1. Build and deploy static files

npm ci
npm run build

aws s3 sync ./static s3://"$S3_BUCKET" --delete

# 2. Build and push Lambda container image

TAG=$(git rev-parse --short HEAD)

echo "$TAG" > version

docker buildx build . \
    --tag "$ECR_REPO:$TAG" \
    --platform linux/arm64 \
    --push \
    --provenance=false

# 3. Update Lambda function code

SHA_256=$(aws ecr describe-images \
            --repository-name lambda/ipv6test \
            --image-ids imageTag="$TAG" \
            --query 'imageDetails[0].imageDigest' \
            --output text)

wait_for_status() {
    local description="$1"
    local max_attempts=30
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        local current_status
        current_status=$(aws lambda get-function-configuration \
            --function-name "$FUNCTION_NAME" \
            --query "LastUpdateStatus" \
            --output text 2>/dev/null)
        if [ "$current_status" = "Successful" ]; then
            echo "$description successful"
            return 0
        fi
        if [ "$current_status" = "Failed" ]; then
            echo "$description failed"
            return 1
        fi
        echo "Waiting for $description (attempt $((attempt+1))/$max_attempts)"
        sleep 10
        ((attempt++))
    done

    echo "Timeout waiting for update"
    return 1
}

aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --image-uri "$ECR_REPO@$SHA_256"

wait_for_status 'Function code update' || exit 1

# 4. Create version

VERSION=$(aws lambda publish-version \
            --function-name "$FUNCTION_NAME" \
            --description "New deployment" \
            --query 'Version' \
            --output text)

wait_for_status "Version $VERSION" || exit 1

# 5. Assign alias

aws lambda update-alias \
    --function-name "$FUNCTION_NAME" \
    --function-version "$VERSION" \
    --name main
