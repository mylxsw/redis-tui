package main

import (
	"flag"
	"fmt"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Cluster  bool
	Debug    bool
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
	flag.BoolVar(&config.Debug, "vvv", false, "Enable debug mode")

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version and exit")

	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s\nGitCommit: %s\n", Version, GitCommit)

		return
	}

	outputChan := make(chan OutputMessage, 10)

	client := NewRedisClient(config, outputChan)
	if err := NewRedisGli(client, 100, Version, GitCommit, outputChan, config).Start(); err != nil {
		panic(err)
	}
}
