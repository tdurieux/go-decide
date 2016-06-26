package main

import (
	"github.com/tdurieux/go-decide/decide"
	"fmt"
	"encoding/json"
	"os"
	"flag"
	"io/ioutil"
	"strings"
	"path"
)

func getInput(inputPath string) (decide.INPUT, error) {
	var input decide.INPUT
	configFile, err := os.Open(inputPath)
	defer configFile.Close()
	if err != nil {
		return input, err
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&input); err != nil {
		return input, err
	}
	return input, nil
}

func serializeDecision(decision decide.Decide) []byte {
	strOutput, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		println("err output", err.Error())
		return nil
	}
	return strOutput
}

func execute(filePath string, outputDir string) decide.Decide {
	decision := decide.Decide{}

	input, err := getInput(filePath)
	if err != nil {
		println("unable to get the input file", err.Error())
		return decision
	}

	err = decision.Decide(input)
	if outputDir != "" {
		os.MkdirAll(outputDir, 0700)
		outputFile := path.Join(outputDir, path.Base(filePath))
		ioutil.WriteFile(outputFile, serializeDecision(decision), 0644)
	}
	return decision
}

func main() {
	filePath := flag.String("input", "", "the path to the input")
	outputPath := flag.String("output", "", "the path to the output")
	flag.Parse()

	if *filePath == "" {
		flag.Usage()
		return
	}
	f, err := os.Open(*filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, _ := ioutil.ReadDir(*filePath)
		for _, f := range files {
			name := f.Name()
			if (strings.Contains(name, ".json")) {
				ff := *filePath + ""
				file := path.Join(ff, name)
				fmt.Print(name)
				decide := execute(file, *outputPath)
				fmt.Println(" " + decide.Launch)
			}
		}
	case mode.IsRegular():
		decide := execute(*filePath, *outputPath)
		fmt.Println(decide.Launch)
	}
}
