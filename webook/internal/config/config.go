package config

type config struct {
	DB    DbConfig
	Redis RedisConfig
}

type DbConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}
