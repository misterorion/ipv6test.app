#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Function to wait for a specific status with retries
wait_for_status() {
    local function_name="$1"
    local status_query="$2"
    local description="$3"
    local max_attempts=${4:-60}
    local sleep_interval=${5:-10}

    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        local current_status=$(aws lambda get-function-configuration \
            --function-name "$function_name" \
            --query "$status_query" \
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
        sleep "$sleep_interval"
        ((attempt++))
    done

    # Timeout
    echo "Timeout waiting for $description"
    return 1
}

# Main deployment script
main() {
    aws lambda update-function-code \
        --function-name "$FUNCTION_NAME" \
        --image-uri "$FULL_IMAGE_URI"

    wait_for_status "$FUNCTION_NAME" 'LastUpdateStatus' 'Function code update' || exit 1

    local VERSION=$(aws lambda publish-version \
        --function-name "$FUNCTION_NAME" \
        --description "New deployment with updated image" \
        --query 'Version' \
        --output text)

    wait_for_status "$FUNCTION_NAME" "LastUpdateStatus" "Version $VERSION" || exit 1

    aws lambda update-alias \
        --function-name "$FUNCTION_NAME" \
        --function-version "$VERSION" \
        --name main

    echo "Deployment completed successfully. New version $VERSION is now live."
}

main "$@"