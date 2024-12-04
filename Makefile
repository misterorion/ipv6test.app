build:
	aws ecr batch-delete-image --repository-name lambda/ipv6test --image-ids imageTag=latest
	# tag=$(git rev-parse --short HEAD)-$(date +"%Y%m%d-%H%M%S")
	# docker buildx build --platform linux/arm64 -t 733051452450.dkr.ecr.us-east-2.amazonaws.com/lambda/ipv6test:$tag . --push
	# docker tag 733051452450.dkr.ecr.us-east-2.amazonaws.com/lambda/ipv6test:$tag 733051452450.dkr.ecr.us-east-2.amazonaws.com/lambda/ipv6test:latest
	# docker push 733051452450.dkr.ecr.us-east-2.amazonaws.com/lambda/ipv6test:latest