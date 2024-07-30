package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func validateURL(url string) (string, error) {
	temp := strings.ToLower(strings.TrimSpace(url))
	if !strings.HasPrefix(temp, "http://") && !strings.HasPrefix(temp, "https://") {
		url = "https://" + url
	}
	if len(url) == 0 {
		return "", fmt.Errorf("url is empty")
	}
	return url, nil
}

func get(url string) ([]byte, error) {

	url, err := validateURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to validate URL: %v", err)
	}

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

func getBodyInnerText(htmlContent []byte) (string, error) {
	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}

	var bodyText string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyText = extractText(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if bodyText == "" {
		return "", fmt.Errorf("no body element found")
	}

	return bodyText, nil
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c)
	}
	return text
}

func GetContent(url string) ([]string, error) {
	body, err := get(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get content: %v", err)
	}

	content, err := getBodyInnerText(body)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get body text: %v", err)
	}

	paragraphs := make([]string, 0)
	sentences := strings.Split(content, ".")
	for _, sentence := range sentences {
		trimmedSentence := strings.TrimSpace(sentence)
		if trimmedSentence != "" {
			paragraphs = append(paragraphs, trimmedSentence+".")
		}
	}

	return paragraphs, nil
}
