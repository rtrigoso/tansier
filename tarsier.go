package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"math/rand"
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
	regATag1 = regexp.MustCompile(`(?mi)<a.+?href\=\"(.+?)\".*?\>`)
	regATag2 = regexp.MustCompile(`(?mi)<a.+?href\=\"(.+?)\".*?\>(.*?)\<\/a\>`)
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

	body, err := getBody(args[0])
	if err != nil {
		return
	}

	if *r {
		p := bluemonday.NewPolicy()
		p.AllowAttrs("href").OnElements("a").AllowURLSchemes("http", "https")
		body = p.Sanitize(string(body))

		links := regATag2.FindAllString(body, -1)
		if len(links) <= 0 {
			fmt.Println("Error: tarsier was not able to find parsable links within the passed url")
			flag.Usage()
			os.Exit(0)
		}

		articleUrl := regATag2.ReplaceAllString(links[rand.Intn(len(links))], "$1")

		fmt.Printf("Reading link %s\n", articleUrl)
		body, err = getBody(articleUrl)
		if err != nil {
			return
		}
	}

	article := getArticle(body)

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

func getBody(site string) (string, error) {
	parseUrl, err := url.Parse(site)
	if err != nil {
		return "", err
	}

	if !parseUrl.IsAbs() {
		parseUrl.Scheme = "https"
	}

	resp, err := http.Get(parseUrl.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
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
	return strings.TrimSpace(content.String())
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
	content = regATag2.ReplaceAllString(content, "$2 (<blue>$1</>)")
	content = regATag1.ReplaceAllString(content, "<blue>$1</>")

	return content
}
