package main

import (
	"github.com/urfave/cli/v2"
)

func makeApp() *cli.App {
	return &cli.App{
		Name:  "hans",
		Usage: "an app for managing and generating corpuses, used internally by OmniLotl",
		Commands: []*cli.Command{
			{
				Name:   "gendb",
				Usage:  "<input-file> <output-dir> generates a KVS DB (level) using the input file",
				Action: genDBMain,
			},
			{
				Name:   "variate",
				Usage:  "<word> takes a standard hinglish word and generates variations for it",
				Action: variateMain,
			},
			{
				Name:   "wordindex",
				Usage:  "<input-file> <output-dir> counts word in the given file and produces a KVS",
				Action: wordIndexMain,
			},
		},
	}
}
