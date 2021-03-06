package main

import (
	"log"
	"html/template"
	"os"
	"flag"
	"io/ioutil"
	"strings"
	"path"
	"encoding/json"
	"github.com/tdurieux/go-decide/decide"
	"fmt"
	"bufio"
	"sort"
	"strconv"
)

const tplResults = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta http-equiv="x-ua-compatible" content="ie=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="css/pure-min.css">
		<link rel="stylesheet" type="text/css" href="css/style.css">
	</head>
	<body>
		<h1>Decisions</h1>
		<table class="summary pure-table">
			<thead>
				<tr>
					<th>Input</th>
					<th>Reference</th>
					<th>Average</th>
					<th>1</th>
					<th>2</th>
					<th>3</th>
					<th>4</th>
					<th>5</th>
					<th>6</th>
					<th>7</th>
				</tr>
			</thead>
			<tbody>{{range $i, $v := .Items}}
				<tr>
					<td>{{ $i }}</td>{{range $v}}
					<td class="{{ . }}">{{ . }}</td>{{end}}
				</tr>{{end}}
			</tbody>
		</table>

	</body>
</html>`

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta http-equiv="x-ua-compatible" content="ie=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="css/pure-min.css">
		<link rel="stylesheet" type="text/css" href="css/style.css">
	</head>
	<body>
		<h1>Decisions</h1>
		<table class="summary pure-table">
			<thead>
				<tr><th>Input</th><th>is to launch?</th></tr>
			</thead>
			<tbody>{{range .Items}}
				<tr>
					<td><a href="{{ .File }}.html">{{ .File }}</a></td>
					<td class="{{if eq .Decision.Launch "YES"}}yes{{else}}no{{end}}">{{ .Decision.Launch }}</td>
				</tr>{{end}}
			</tbody>
		</table>

	</body>
</html>`

const tplDetails = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta http-equiv="x-ua-compatible" content="ie=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="css/pure-min.css">
		<link rel="stylesheet" type="text/css" href="css/style.css">
	</head>
	<body>
		<h1>{{.Title}} - {{ .Item.Decision.Launch }}</h1>
		<h2>CMV</h2>
		<table class="summary pure-table">
			<thead>
				<tr>{{range $i, $e := .Item.Decision.CMV}}
					<th>{{ $i }}</th>{{end}}
				</tr>
			</thead>
			<tbody>
				<tr>{{range .Item.Decision.CMV}}
					<td class="{{if .}}yes{{else}}no{{end}}">{{if .}}V{{else}}X{{end}}</td>{{end}}
				</tr>
			</tbody>
		</table>
		<h2>FUV</h2>
		<table class="summary pure-table">
			<thead>
				<tr>{{range $i, $e := .Item.Decision.FUV}}
					<th>{{ $i }}</th>{{end}}
				</tr>
			</thead>
			<tbody>
				<tr>{{range .Item.Decision.FUV}}
					<td class="{{if .}}yes{{else}}no{{end}}">{{if .}}V{{else}}X{{end}}</td>{{end}}
				</tr>
			</tbody>
		</table>
		<h2>PUM</h2>
		<table class="summary pure-table">
			<thead>
				<tr>
					<th></th>{{range $i, $e := .Item.Decision.PUM}}
					<th>{{ $i }}</th>{{end}}
				</tr>
			</thead>
			<tbody>{{range $i, $e :=  .Item.Decision.PUM}}
				<tr>
					<th>{{ $i }}</th>{{range $e}}
					<td class="{{if .}}yes{{else}}no{{end}}">{{if .}}V{{else}}X{{end}}</td>{{end}}
				</tr>{{end}}
			</tbody>
		</table>
	</body>
</html>`

type Results map[string][]string

type FileDecide struct {
	File string
	Decision decide.Decide
}
type ById []FileDecide

func fileToInt(file string) int {
	i, _ := strconv.Atoi(strings.Replace(strings.Replace(file, ".json", "", 1), "input", "", 1))
	return i
}
func (a ById) Len() int           { return len(a) }
func (a ById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ById) Less(i, j int) bool {
	return fileToInt(a[i].File) < fileToInt(a[j].File)
}

func getResults(outputDir string) Results {
	file := path.Join(outputDir, "results-1001.json")

	executionOutput, _ := os.Open(file)
	defer executionOutput.Close()
	results := Results{}
	jsonParser := json.NewDecoder(executionOutput)
	err := jsonParser.Decode(&results);
	if (err != nil) {
		log.Fatal(err)
	}
	nbValid := make([]int, 9)
	for k, v := range results {
		newValue := make([]string, len(v) + 1)
		refVal := v[0]
		newValue[0] = refVal
		isEquals := true
		countYes := 0
		for i := 1; i < len(v); i++ {
			newValue[i + 1] = v[i]
			if (isEquals) {
				isEquals = refVal == v[i]
			}
			if v[i] == "yes" {
				countYes++
			}
		}

		if isEquals {
			delete(results, k)
			newValue[1] = refVal
		} else {
			if (countYes > (len(v) - 1)/ 2) {
				newValue[1] = "yes"
			} else if (countYes > (len(v) - 1)/ 2) {
				newValue[1] = "middle"
			} else {
				newValue[1] = "no"
			}
			results[k] = newValue
		}
		for i := 0; i < len(newValue); i++ {
			if newValue[i] == refVal {
				nbValid[i] ++
			}
		}
	}
	correct := make([]string, 9)
	for i, _ := range correct {
		correct[i] = fmt.Sprintf("%d", nbValid[i])
	}
	results["Nb correct"] = correct
	return results;
}

func getInputs(outputDir string) []FileDecide {
	files, _ := ioutil.ReadDir(outputDir)
	output := make([]FileDecide, len(files))
	for i, f := range files {
		name := f.Name()
		if (strings.Contains(name, ".json") && strings.Contains(name, "input")) {
			file := path.Join(outputDir, name)

			executionOutput, _ := os.Open(file)
			decision := decide.Decide{}
			jsonParser := json.NewDecoder(executionOutput)
			err := jsonParser.Decode(&decision);
			if err == nil {
				output[i] = FileDecide{strings.Replace(name, ".json", "", 1), decision}
			} else {
				fmt.Println(err)
			}
			executionOutput.Close()
		}
	}
	return output;
}

func main() {
	outputPath := flag.String("output", "", "the path to the output")
	flag.Parse()

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(tplResults)
	check(err)

	results := getResults(*outputPath)
	dataResults := struct {
		Title string
		Items Results
	}{
		Title: "Decides Results",
		Items: results,
	}
	f, err := os.Create(path.Join(*outputPath, "..", "results.html"))
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	err = t.Execute(w, dataResults)
	w.Flush()
	check(err)

	t, err = template.New("webpage").Parse(tpl)
	check(err)

	inputs := getInputs(*outputPath)
	sort.Sort(ById(inputs))
	data := struct {
		Title string
		Items []FileDecide
	}{
		Title: "Decide",
		Items: inputs,
	}
	f, err = os.Create(path.Join(*outputPath, "..", "index.html"))
	check(err)
	defer f.Close()
	w = bufio.NewWriter(f)
	err = t.Execute(w, data)
	w.Flush()
	check(err)

	for _, input := range inputs {
		t, err = template.New("webpage").Parse(tplDetails)
		check(err)
		dataDetail := struct {
			Title string
			Item FileDecide
		}{
			Title: "Decide - " + input.File,
			Item: input,
		}

		f, err = os.Create(path.Join(*outputPath, "..", input.File + ".html"))
		check(err)
		w = bufio.NewWriter(f)
		err = t.Execute(w, dataDetail)
		w.Flush()
		f.Close()
		check(err)
	}
}
