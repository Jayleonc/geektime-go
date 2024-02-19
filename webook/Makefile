.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build -tags=dev -o webook .
	@docker rmi jayleonc/webook:v0.0.1
	@docker build -t jayleonc/webook:v0.0.1 .

.PHONY: mock
mock:
	mockgen -source=./internal/service/user.go  -destination=./internal/service/mocks/user_mock.go; \
	mockgen -source=./internal/service/article.go  -destination=./internal/service/mocks/article_mock.go; \
	mockgen -source=./internal/service/interactive.go  -destination=./internal/service/mocks/interactive_mock.go; \
  	mockgen -source=./internal/service/code.go  -destination=./internal/service/mocks/code_mock.go; \
  	mockgen -source=./internal/repository/user.go  -destination=./internal/repository/mocks/user_mock.go; \
  	mockgen -source=./internal/repository/code.go  -destination=./internal/repository/mocks/code_mock.go; \
  	mockgen -source=./internal/repository/article.go  -destination=./internal/repository/mocks/article_mock.go; \
  	mockgen -source=./internal/repository/dao/user.go  -destination=./internal/repository/dao/mocks/user_mock.go; \
  	mockgen -source=./internal/repository/cache/user.go  -destination=./internal/repository/cache/mocks/user_mock.go; \
  	mockgen -source=./internal/repository/cache/code.go  -destination=./internal/repository/cache/mocks/code_mock.go; \
  	mockgen -destination=./internal/repository/cache/redismocks/cmd_mock.go github.com/redis/go-redis/v9 Cmdable; \
  	mockgen -source=./internal/repository/article_author.go  -destination=./internal/repository/mocks/article_author_mock.go; \
  	mockgen -source=./internal/repository/dao/article_author.go  -destination=./internal/repository/dao/mocks/article_author_mock.go; \
  	mockgen -source=./internal/repository/dao/article_reader.go  -destination=./internal/repository/dao/mocks/article_reader_mock.go; \
