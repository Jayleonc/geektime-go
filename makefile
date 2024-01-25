.PHONY: mock
mock:
	mockgen -source=./webook/internal/service/user.go -package=svcmocks -destination=./webook/internal/service/mocks/user_mock.go; \
 	mockgen -source=./webook/internal/service/user.go  -destination=./webook/internal/service/mocks/user_mock.go; \
  	mockgen -source=./webook/internal/service/code.go  -destination=./webook/internal/service/mocks/code_mock.go; \
  	mockgen -source=./webook/internal/repository/user.go  -destination=./webook/internal/repository/mocks/user_mock.go; \
  	mockgen -source=./webook/internal/repository/code.go  -destination=./webook/internal/repository/mocks/code_mock.go; \
  	mockgen -source=./webook/internal/repository/dao/user.go  -destination=./webook/internal/repository/dao/mocks/user_mock.go; \
  	mockgen -source=./webook/internal/repository/cache/user.go  -destination=./webook/internal/repository/cache/mocks/user_mock.go; \
  	mockgen -source=./webook/internal/repository/cache/code.go  -destination=./webook/internal/repository/cache/mocks/code_mock.go; \
  	mockgen -destination=./webook/internal/repository/cache/redismocks/cmd_mock.go github.com/redis/go-redis/v9 Cmdable; \
