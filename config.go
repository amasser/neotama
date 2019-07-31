package main

import (
	"github.com/akamensky/argparse"
	"os"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"log"
)

type JobConfig struct {
	url string
	threads int
	pConfig ParserConfig
}

type ParserConfig struct {
	Options struct {
		EnableDescription bool
		TimeFormat string
	}
	Regex struct {
		LineMatch string
		PathGroup int
		TimeGroup int
		SizeGroup int
		DescriptionGroup int
	}
	CompiledRegexp *regexp.Regexp
}

func readParserConfig(filename string) (result ParserConfig) {
	fileContents, fileErr := ioutil.ReadFile(filename)
	if fileErr != nil {
		log.Fatalf("Could not read parser config file \"%s\".", filename)
	}
	jsonErr := json.Unmarshal(fileContents, &result)
	if jsonErr != nil {
		log.Fatalf("Parser config file \"%s\" is not valid JSON.", filename)
	}
	compRegexp, regexpErr := regexp.Compile(result.Regex.LineMatch)
	if regexpErr != nil {
		log.Fatalf("Invalid regex in parser config file.")
	}
	result.CompiledRegexp = compRegexp
	return
}

func ReadConfig() (config JobConfig) {
	parser := argparse.NewParser("", "Safely and quickly crawls a directory listing and outputs a pretty tree.")

	url := parser.String("u", "url", &argparse.Options{Required: true, Help: "URL to crawl"})
	threads := parser.Int("t", "threads", &argparse.Options{Required: false, Help: "Maximum number of network threads to open at once", Default: 10})
	configFile := parser.String("c", "config", &argparse.Options{Required: false, Help: "Config file to use for parsing the directory listing."})

	parser.Parse(os.Args)

	config.url = *url
	config.threads = *threads
	config.pConfig = readParserConfig((*configFile))

	return
}
