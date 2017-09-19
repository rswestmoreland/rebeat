package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/rswestmoreland/rebeat/beater"
)

func main() {
	err := beat.Run("rebeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
