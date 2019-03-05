package main

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/repr"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// parseFiles := os.Args
	config := ParseFile("files/zakay.txt", nil)

	config = config.splitAndAssociateChildren()

	repr.Println(config, repr.Indent("  "), repr.OmitEmpty(true))

	mapped, err := json.Marshal(config.toMap())

	check(err)

	fmt.Println(string(mapped))
}
