//go:build k8s

package config

var Config = config{
	DB: DbConfig{
		DSN: "root:root@tcp(webook-mysql:3308)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:6380",
	},
}
