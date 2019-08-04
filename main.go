package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
var pythonFile string

var workingDirectory string

func init() {
	flag.StringVar(&outFile, "o", "", "Output filename")
	flag.StringVar(&inFile, "i", "", "Input filename")
	flag.StringVar(&pythonFile, "p", "", "Python generator file")
}

func printUsage() {
	fmt.Println("usage: " + os.Args[0] + " -i inFile [-o outFile] [-p pythonFile]")
}

func main() {
	flag.Parse()

	if inFile == "" || len(os.Args) < 2 {
		stderr.Println("Need a file to parse")
		printUsage()
		flag.PrintDefaults()
		os.Exit(1)
	}

	currentPath, err := os.Getwd()

	check(err)

	workingDirectory = currentPath

	path, err := filepath.Abs(inFile)

	check(err)

	pathParts := strings.Split(path, "/")

	pathParts = pathParts[:len(pathParts)-1]

	path = strings.Join(pathParts, "/")

	err = os.Chdir(path)

	check(err)

	inFilePathParts := strings.Split(inFile, "/")

	inFileName := inFilePathParts[len(inFilePathParts)-1]

	config := ParseFile(inFileName, nil)

	mapped := config.Transform()

	marshaled, err := json.Marshal(mapped)

	check(err)

	if outFile == "" {
		fmt.Println(string(marshaled))
	} else {
		ioutil.WriteFile(outFile, marshaled, 0644)
	}
}
