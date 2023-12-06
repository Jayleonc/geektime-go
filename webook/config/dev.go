//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:jayleonc@tcp(175.178.58.198:33306)/geektime",
	},
	Redis: RedisConfig{
		Addr:     "175.178.58.198:16379",
		Password: "jayleonc",
	},
}
