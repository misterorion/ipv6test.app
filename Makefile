
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

dev:
	@npm run dev

deploy:
	@git add .
	@git commit -m "Updates"
	@git push origin main
	@aws codebuild start-build --project-name ipv6test-app --query 'build.buildStatus'
