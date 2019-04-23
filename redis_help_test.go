package main

import (
	"testing"
)

func TestRedisHelpMatch(t *testing.T) {
	if !RedisHelpMatch("client get", func(help RedisHelp) {
		if help.Command != "CLIENT GETNAME" {
			t.Error("test failed")
		}
	}) {
		t.Error("test failed")
	}
}

func TestRedisMatchedCommands(t *testing.T) {
	cmds := RedisMatchedCommands("cluster re")
	if len(cmds) != 3 {
		t.Error("test failed")
	}
}
