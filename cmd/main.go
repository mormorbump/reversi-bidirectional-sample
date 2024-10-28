package main

import (
	"kazuki.matsumoto/reversi/client"
	"os"
)

func main() {
	os.Exit(client.NewReversi().Run())
}
