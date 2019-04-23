package config

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Cluster  bool
	Debug    bool
}