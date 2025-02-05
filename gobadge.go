package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Threshold struct {
	yellow int
	green  int
}

type Params struct {
	label     string
	threshold Threshold
	color     string
	value     string
	link      string
}

func main() {
	source := flag.String("filename", "output.out", "File containing the tests output")
	label := flag.String("text", "Coverage", "Text on the left side of the badge")
	yellowThreshold := flag.Int("yellow", 30, "At what percentage does the badge becomes yellow instead of red")
	greenThreshold := flag.Int("green", 70, "At what percentage does the badge becomes green instead of yellow")
	color := flag.String("color", "", "Color of the badge - green/yellow/red")
	target := flag.String("target", "coverage.svg", "Target file")
	value := flag.String("value", "", "Text on the right side of the badge")
	link := flag.String("link", "", "Link the badge goes to")

	flag.Parse()

	params := &Params{
		*label,
		Threshold{*yellowThreshold, *greenThreshold},
		*color,
		*value,
		*link,
	}

	err := generateBadge(*source, *target, params)

	if err != nil {
		log.Fatal(err)
	}
}

func generateBadge(source string, target string, params *Params) error {
	var coverage string
	var err error

	if params.value != "" {
		coverage = params.value
	} else {
		coverage, err = retrieveTotalCoverage(source)
	}

	if err != nil {
		return err
	}

	badgeColor := setColor(coverage, params.threshold.yellow, params.threshold.green, params.color)
	err = saveSvg(target, coverage, params.label, badgeColor)

	if err != nil {
		return err
	}

	fmt.Println("\033[0;36mGoBadge: Coverage badge updated to " + coverage + " in " + target + "\033[0m")

	return nil
}

func setColor(coverage string, yellowThreshold int, greenThreshold int, color string) string {
	coverageNumber, _ := strconv.ParseFloat(strings.Replace(coverage, "%", "", 1), 4)
	if color != "" {
		return color
	}
	if coverageNumber >= float64(greenThreshold) {
		return "brightgreen"
	}
	if coverageNumber >= float64(yellowThreshold) {
		return "yellow"
	}
	return "red"
}

func retrieveTotalCoverage(filename string) (string, error) {
	// Read coverage file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("\033[1;31mGoBadge: Error while opening the coverage file\033[0m")
		return "", err
	}
	defer file.Close()

	// split content by words and grab the last one (total percentage)
	b, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("\033[1;31mGoBadge: Error while reading the coverage file\033[0m")
		return "", err
	}
	words := strings.Fields(string(b))
	last := words[len(words)-1]

	return last, nil
}

func saveSvg(target string, coverage string, label string, color string) error {
	encodedLabel := url.QueryEscape(label)
	encodedCoverage := url.QueryEscape(coverage)
	urlx := fmt.Sprintf(`https://img.shields.io/badge/%s-%s-%s`, encodedLabel, encodedCoverage, color)

	response, e := http.Get(urlx)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	//open a file for writing
	file, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
