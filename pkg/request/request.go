package request

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

var target string

type HTMLElement interface {
	Render() template.HTML
}
type Paragraph struct {
	Content string
}

func (p Paragraph) Render() template.HTML {
	sentenceElements := ""
	sentences := strings.Split(p.Content, ".")
	for _, sentence := range sentences {
		trimmedSentence := strings.TrimSpace(sentence)
		if trimmedSentence != "" {
			sentenceElements += fmt.Sprintf("<p>%s.</p>", trimmedSentence)
		}
	}
	return template.HTML(sentenceElements)

}

type Image struct {
	Source          string
	AlternativeText string
}

func (i Image) Render() template.HTML {
	var src string
	if hasProtocol(i.Source) {
		src = i.Source
	} else {
		root, err := getUrlRoot(target)
		if err == nil {
			src = root + i.Source
		} else {
			src = "/na"
		}
	}
	return template.HTML(fmt.Sprintf("<img src=\"%s\" alt=\"%s\">", src, i.AlternativeText))
}

func validateURL(url string) (string, error) {
	if !hasProtocol(url) {
		url = "https://" + url
	}
	if len(url) == 0 {
		return "", fmt.Errorf("url is empty")
	}
	return url, nil
}

func getUrlRoot(fullURL string) (string, error) {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("URL must contain a valid scheme and host")
	}

	root := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return root, nil
}

func hasProtocol(url string) bool {
	url = strings.ToLower(strings.TrimSpace(url))
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func getResponseBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func getInnerHTMLElements(htmlContent []byte) ([]HTMLElement, error) {
	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return []HTMLElement{}, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var elements []HTMLElement
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			elements = extractHTMLElements(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if len(elements) == 0 {
		return []HTMLElement{}, fmt.Errorf("no body element found")
	}

	return elements, nil
}

func extractHTMLElements(n *html.Node) []HTMLElement {
	var elements []HTMLElement
	paragraph := Paragraph{
		Content: "",
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script") {
			return
		} else if n.Type == html.TextNode {
			content := strings.TrimSpace(n.Data)
			if content != "" {
				paragraph.Content += content
			}
		} else if n.Type == html.ElementNode && n.Data == "img" {
			if paragraph.Content != "" {
				elements = append(elements, paragraph)
				paragraph = Paragraph{
					Content: "",
				}
			}
			var src, alt string
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					src = attr.Val
				} else if attr.Key == "alt" {
					alt = attr.Val
				}
			}
			elements = append(elements, Image{Source: src, AlternativeText: alt})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	if paragraph.Content != "" {
		elements = append(elements, paragraph)
	}

	return elements
}

func GetContent(url string) ([]HTMLElement, error) {
	url, err := validateURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to validate URL: %v", err)
	}

	target = url

	body, err := getResponseBody(url)
	if err != nil {
		return []HTMLElement{}, fmt.Errorf("failed to get response body: %v", err)
	}

	content, err := getInnerHTMLElements(body)
	if err != nil {
		return []HTMLElement{}, fmt.Errorf("failed to get inner HTML elements: %v", err)
	}

	return content, nil
}
