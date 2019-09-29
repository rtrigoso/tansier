package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"github.com/microcosm-cc/bluemonday"
)

var (
	r   = flag.Bool("r", false, "")
	err error

	regP     = regexp.MustCompile(`(?mi)(\<p\>.+?\<\/p\>)`)
	regATag  = regexp.MustCompile(`(?mi)<a.+?href\=\"(.+?)\".*?\>`)
	regALink = regexp.MustCompile(`(?mi)<a.+?href\=\"(.+?)\".*?\>(.*?)\<\/a\>`)
)

var usage = `Usage: tarsier [options...] url

Options:
	-r	Selects a random link from within the url provided and outputs that instead
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
	}
	flag.Parse()

	defer func() {
		if err != nil {
			fmt.Printf("error running tarsier: %v", err)
			os.Exit(1)
		}
	}()

	args := flag.Args()
	if len(args) <= 0 {
		flag.Usage()
		os.Exit(0)
	}

	site := args[0]
	parseUrl, err := url.Parse(site)
	if err != nil {
		return
	}

	if !parseUrl.IsAbs() {
		parseUrl.Scheme = "https"
	}

	resp, err := http.Get(parseUrl.String())
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	article := strings.TrimSpace(getArticle(string(body)))
	if article == "" {
		fmt.Println("Error: tarsier was not able to find an article in the provided website url")
		flag.Usage()
		os.Exit(0)
	}

	p := buildParsePolicy()
	html := p.Sanitize(string(article))

	for _, match := range regP.FindAllString(html, -1) {
		p := createParagraph(match)
		color.Println(p)
	}
}

func buildParsePolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "b", "strong", "code", "em")
	p.AllowAttrs("href").OnElements("a")
	return p
}

func getArticle(s string) string {
	var content strings.Builder
	for _, match := range regP.FindAllString(s, -1) {
		content.WriteString(match)
		content.WriteString("\n\n")
	}
	return content.String()
}

func createParagraph(s string) string {
	ps := regP.FindAllString(s, -1)
	if len(ps) <= 0 {
		return ""
	}

	content := ps[0]
	content = html.UnescapeString(content)
	content = strings.Replace(content, "<em>", "<red>", -1)
	content = strings.Replace(content, "</em>", "</>", -1)
	content = strings.Replace(content, "<strong>", "<bold>", -1)
	content = strings.Replace(content, "</strong>", "</>", -1)
	content = strings.Replace(content, "<b>", "<bold>", -1)
	content = strings.Replace(content, "</b>", "</>", -1)
	content = strings.Replace(content, "<code>", "<green>", -1)
	content = strings.Replace(content, "</code>", "</>", -1)
	content = strings.Replace(content, "<p>", "", -1)
	content = strings.Replace(content, "</p>", "", -1)
	content = regALink.ReplaceAllString(content, "$2 (<blue>$1</>)")
	content = regATag.ReplaceAllString(content, "<blue>$1</>")

	return content
}
