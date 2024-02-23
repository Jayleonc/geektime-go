package error

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrKeyNotExist = redis.Nil
var ErrRecordNotFound = gorm.ErrRecordNotFound
