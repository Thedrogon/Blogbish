package config

type Config struct {
	Server  ServerConfig
	Storage StorageConfig
	Cache   CacheConfig
}

type ServerConfig struct {
	Port         string
	AllowOrigins []string
}

type StorageConfig struct {
	Provider     string // "s3" or "local"
	BucketName   string
	Region       string
	MaxFileSize  int64
	AllowedTypes []string
}

type CacheConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
}
