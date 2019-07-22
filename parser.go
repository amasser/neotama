package main

import (
	"math"
	"regexp"
	"strings"
	"strconv"
	"html"
	"net/url"
	"time"
)

var sizeMatcher *regexp.Regexp

func init() {
	sizeMatcher = regexp.MustCompile(`(?m)^([0-9\.]*)(.*)$`)
}

func parseFilesize(filesize string) (bytes int64) {
	matches := sizeMatcher.FindStringSubmatch(filesize)
	// TODO: Error handling
	var multiplier float64 = 1000
	if strings.ContainsAny(matches[2], "iI") {
		multiplier = 1024
	}
	switch {
		case strings.ContainsAny(matches[2], "kK"): multiplier = math.Pow(multiplier, 1)
		case strings.ContainsAny(matches[2], "mM"): multiplier = math.Pow(multiplier, 2)
		case strings.ContainsAny(matches[2], "gG"): multiplier = math.Pow(multiplier, 3)
		case strings.ContainsAny(matches[2], "tT"): multiplier = math.Pow(multiplier, 4)
		case strings.ContainsAny(matches[2], "pP"): multiplier = math.Pow(multiplier, 5)
		case strings.ContainsAny(matches[2], "zZ"): multiplier = math.Pow(multiplier, 6)
		default: multiplier = 1
	}
	// TODO: Error handling
	floatBytes, _ := strconv.ParseFloat(matches[1], 64)
	return int64(floatBytes * multiplier)
}

func cleanHtml(input string) (result string) {
	result = html.UnescapeString(input)
	result = strings.TrimSpace(result)
	return
}

type DirListEntry struct {
	path string
	time string
	size string
	description string
}

func splitDirListEntry(html []string, pConfig ParserConfig) (entry DirListEntry) {
	entry.path = html[pConfig.Regex.PathGroup]
	entry.time = html[pConfig.Regex.TimeGroup]
	entry.size = html[pConfig.Regex.SizeGroup]
	entry.description = html[pConfig.Regex.DescriptionGroup]
	return
}

func lastChar(s string) (c string) {
	// TODO: Error handling
	return s[len(s)-1:]
}

func parseDirListEntry(html []string, parentUrl string, pConfig ParserConfig) (node WebNode, skip bool) {
	e := splitDirListEntry(html, pConfig)
	cleanTime := cleanHtml(e.time)
	if cleanTime == "" {
		skip = true
		return
	}
	node.path = parentUrl + e.path
	node.name, _ = url.QueryUnescape(e.path)
	if (lastChar(node.path) == "/") {
		node.nodeType = directory
		node.nodeStatus = pending
	} else {
		node.nodeType = file
		node.nodeStatus = done
	}
	time, _ := time.Parse(pConfig.Options.TimeFormat, cleanTime)
	node.time = &time
	if pConfig.Options.EnableDescription {
		node.description = cleanHtml(e.description)
	} else {
		node.description = ""
	}
	node.size = parseFilesize(cleanHtml(e.size))
	return
}

func ParseDirList(html, parentUrl string, pConfig ParserConfig) (nodes []WebNode) {
	dirListEntries := pConfig.CompiledRegexp.FindAllStringSubmatch(html, -1)
	// spew.Dump(dirListEntries)
	nodes = make([]WebNode, 0, len(dirListEntries))
	for _, v := range(dirListEntries) {
		node, skip := parseDirListEntry(v, parentUrl, pConfig)
		if !skip {
			nodes = append(nodes, node)
		}
	}
	return
}