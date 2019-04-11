package main

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	// fuzzy match the expected pprof url to extract the port number
	pprofUrlPattern      = regexp.MustCompile(`(?:https?://)?localhost:(\d+)(?:/debug(?:/pprof(?:/goroutine(?:\?debug=\d?)?)?)?)?`)
	goroutineStackHeader = regexp.MustCompile(`goroutine (\d+) \[.*]:\n`)
	originatingFrom      = regexp.MustCompile(`\[originating from goroutine (\d+)]:\n`)
	graphvizFormatter    = strings.NewReplacer(
		"\n", "\\l",
	)
)

func main() {
	url := os.Args[1]
	if matchedURL := pprofUrlPattern.FindStringSubmatch(url); matchedURL != nil {
		url = "http://localhost:"+matchedURL[1]+"/debug/pprof/goroutine?debug=2"
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("failed to download profile: %s", err.Error())
	}

	// have to do yolo text parsing because ancestors aren't yet in the proto format output
	profile, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read profile: %s", err.Error())
	}
	g := gographviz.NewEscape()
	_ = g.SetDir(true)

	// right angle edges
	g.Attrs[gographviz.Splines] = "ortho"

	// increase padding between nodes
	g.Attrs[gographviz.NodeSep] = "1"

	// Set strict will automatically de-duplicate edges for us
	_ = g.SetStrict(true)

	goroutineStacks := strings.Split(string(profile), "\n\n")
	for _, stack := range goroutineStacks {
		createGoroutineNode(stack, g)
	}

	for _, stack := range goroutineStacks {
		addGoroutineStackEdges(stack, g)
	}

	fmt.Println(g.String())
}

func createGoroutineNode(stack string, g *gographviz.Escape) {
	//fmt.Println(stack)
	header := goroutineStackHeader.FindStringSubmatch(stack)
	goroutineNum := header[1]

	//fmt.Println("goroutine:", goroutineNum)
	var text string
	originatingGoroutines := originatingFrom.FindAllStringIndex(stack, -1)
	if len(originatingGoroutines) < 2 {
		text = stack
	} else {
		text = stack[0:originatingGoroutines[1][0]]
	}
	_ = g.AddNode("", goroutineNum, map[string]string{
		"label": graphvizFormatter.Replace(text),
		"shape": "record",
	})
}

func addGoroutineStackEdges(stack string, g *gographviz.Escape) {
	header := goroutineStackHeader.FindStringSubmatch(stack)
	stack = strings.TrimPrefix(stack, header[0])
	goroutineNum := header[1]
	originatingGoroutines := originatingFrom.FindAllStringSubmatch(stack, -1)

	for _, originatingHeader := range originatingGoroutines {
		if !g.IsNode(originatingHeader[1]) {
			// parent has terminated so create a dummy node for them
			_ = g.AddNode("", originatingHeader[1], map[string]string{
				"label": "goroutine "+originatingHeader[1]+" (terminated)",
			})
		}

		_ = g.AddEdge(originatingHeader[1], goroutineNum, true, nil)
		goroutineNum = originatingHeader[1]
	}
}
