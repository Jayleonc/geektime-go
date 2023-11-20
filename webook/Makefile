.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build -tags=dev -o webook .
	@docker rmi jayleonc/webook:v0.0.1
	@docker build -t jayleonc/webook:v0.0.1 .