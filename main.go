package main

import (
	"flag"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Cluster  bool
}

var config = Config{}

var Version string
var GitCommit string

func main() {
	flag.StringVar(&config.Host, "h", "127.0.0.1", "Server hostname")
	flag.IntVar(&config.Port, "p", 6379, "Server port")
	flag.StringVar(&config.Password, "a", "", "Password to use when connecting to the server")
	flag.IntVar(&config.DB, "n", 0, "Database number")
	flag.BoolVar(&config.Cluster, "c", false, "Enable cluster mode")

	flag.Parse()

	client := NewRedisClient(config)

	if err := NewRedisGli(client, 100, Version, GitCommit).Start(); err != nil {
		panic(err)
	}
}