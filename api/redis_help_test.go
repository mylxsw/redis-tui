package api_test

import (
	"github.com/mylxsw/redis-tui/api"
	"testing"
)

func TestRedisHelpMatch(t *testing.T) {
	if !api.RedisHelpMatch("client get", func(help api.RedisHelp) {
		if help.Command != "CLIENT GETNAME" {
			t.Error("test failed")
		}
	}) {
		t.Error("test failed")
	}
}

func TestRedisMatchedCommands(t *testing.T) {
	cmds := api.RedisMatchedCommands("cluster re")
	if len(cmds) != 3 {
		t.Error("test failed")
	}
}
