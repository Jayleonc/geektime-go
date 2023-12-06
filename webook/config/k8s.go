//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:jayleonc@tcp(webook-mysql:13306)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:6379",
	},
}
