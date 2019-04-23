package main

import (
	"flag"
	"fmt"
	"github.com/mylxsw/go-skills/redis-tui/api"
	"github.com/mylxsw/go-skills/redis-tui/config"
	"github.com/mylxsw/go-skills/redis-tui/core"
	"github.com/mylxsw/go-skills/redis-tui/tui"
)

var conf = config.Config{}

var Version string
var GitCommit string

func main() {
	flag.StringVar(&conf.Host, "h", "127.0.0.1", "Server hostname")
	flag.IntVar(&conf.Port, "p", 6379, "Server port")
	flag.StringVar(&conf.Password, "a", "", "Password to use when connecting to the server")
	flag.IntVar(&conf.DB, "n", 0, "Database number")
	flag.BoolVar(&conf.Cluster, "c", false, "Enable cluster mode")
	flag.BoolVar(&conf.Debug, "vvv", false, "Enable debug mode")

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version and exit")

	flag.Parse()

	if len(GitCommit) > 8 {
		GitCommit = GitCommit[:8]
	}

	if showVersion {
		fmt.Printf("Version: %s\nGitCommit: %s\n", Version, GitCommit)

		return
	}

	outputChan := make(chan core.OutputMessage, 10)
	if err := tui.NewRedisTUI(api.NewRedisClient(conf, outputChan), 100, Version, GitCommit, outputChan, conf).Start(); err != nil {
		panic(err)
	}
}
