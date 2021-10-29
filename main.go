package main

import "os"

func main() {
	app := makeApp()
	app.Run(os.Args)
}
