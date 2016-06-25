package main

import (
	"github.com/tdurieux/go-decide/decide"
	"fmt"
	"encoding/json"
	"os"
)

func main() {
	var input decide.INPUT
	configFile, err := os.Open("/home/thomas/goworkspace/src/github.com/tdurieux/go-decide/input/input427.json")
	if err != nil {
		println("opening config file", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&input); err != nil {
		println("parsing config file", err.Error())
	}

	var output = decide.Decide{}
	err = output.Decide(input)
	if err != nil {
		println("err decide", err.Error())
		return
	}

	strOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		println("err output", err.Error())
	}
	fmt.Println(string(strOutput))
}
