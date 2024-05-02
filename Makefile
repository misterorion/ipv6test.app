
clean:
	@rm -rf ./node_modules
	@rm package-lock.json
	@make lockfile

install:
	@npm install

build: install
	@npm run build

lockfile:
	@npm install --package-lock-only

docker: build
	@docker buildx build -t 733051452450.dkr.ecr.us-east-2.amazonaws.com/ipv6test:latest . --push --provenance=false --no-cache

deploy:
	@git add .
	@git commit -m "Updates"
	@git push origin main
	@aws codebuild start-build --project-name ipv6test-app --query 'build.buildStatus'
