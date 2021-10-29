package main

import "github.com/urfave/cli/v2"

const (
	ExitNotEnoughArgs = iota + 1
)

var (
	ErrNotEnoughArgs = cli.Exit("not enough arguments", ExitNotEnoughArgs)
)
