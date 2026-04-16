#!/bin/bash
set -euo pipefail

# 1. Get the Image Digest (Keeps deployment immutable/safe)
echo "Retrieving image digest..."
SHA_256=$(aws ecr describe-images \
    --repository-name "$REPOSITORY_NAME" \
    --image-ids imageTag="$TAG" \
    --query 'imageDetails[0].imageDigest' \
    --output text)

# 2. Update Code and Wait
echo "Updating code and waiting for readiness..."
aws lambda update-function-code \
    --function-name "${FUNCTION_NAME}" \
    --image-uri "${ECR_REPO}@${SHA_256}" > /dev/null

aws lambda wait function-updated --function-name "${FUNCTION_NAME}"

# 3. Publish Version
echo "Publishing version..."
VERSION=$(aws lambda publish-version \
    --function-name "${FUNCTION_NAME}" \
    --description "New deployment" \
    --query 'Version' \
    --output text)

# 4. Update Alias
echo "Updating alias 'main' to version ${VERSION}..."
aws lambda update-alias \
    --function-name "${FUNCTION_NAME}" \
    --function-version "${VERSION}" \
    --name main

echo "Deployment completed successfully."