package main

import (
	"fmt"

	"github.com/GiorgosMarga/satellites/engine"
)

func main() {

	engine := engine.New()

	if err := engine.Start("snapshots"); err != nil {
		fmt.Println(err)
	}
}
