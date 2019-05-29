package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var stderr = log.New(os.Stderr, "", 0)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var outFile string
var inFile string

func init() {
	flag.StringVar(&outFile, "o", "", "Output filename")
	flag.StringVar(&inFile, "i", "", "Input filename")
}

func main() {
	flag.Parse()

	if inFile == "" || len(os.Args) < 2 {
		stderr.Println("Need a file to parse as argument")
		fmt.Println("usage: " + os.Args[0] + " -i inFile [-o outFile]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cwd := path.Dir(inFile)

	setWorkingDirectory(cwd)

	inFile = strings.Replace(inFile, cwd+"/", "", 1)

	config := ParseFile(inFile, nil)

	mapped := config.Transform()

	marshaled, err := json.Marshal(mapped)

	check(err)

	if outFile == "" {
		fmt.Println(string(marshaled))
	} else {
		ioutil.WriteFile(outFile, marshaled, 0644)
	}
}
