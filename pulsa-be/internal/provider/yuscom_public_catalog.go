package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const DefaultYuscomPublicCatalogURL = "https://yuscom.co.id/harga-produk-yuscom"

type YuscomPublicCatalog struct {
	pageURL string
	client  *http.Client
}

func NewYuscomPublicCatalog(pageURL string, timeout time.Duration) *YuscomPublicCatalog {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		pageURL = DefaultYuscomPublicCatalogURL
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &YuscomPublicCatalog{
		pageURL: pageURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *YuscomPublicCatalog) OpenProductCodes(ctx context.Context) (map[string]struct{}, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("katalog publik Yuscom belum dikonfigurasi")
	}
	page, err := c.get(ctx, c.pageURL)
	if err != nil {
		return nil, err
	}
	defer page.Close()

	iframeURL, err := findYuscomCatalogFrame(page, c.pageURL)
	if err != nil {
		return nil, err
	}
	frame, err := c.get(ctx, iframeURL)
	if err != nil {
		return nil, err
	}
	defer frame.Close()

	codes, err := parseOpenYuscomProductCodes(frame)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("katalog publik Yuscom tidak memiliki produk Open")
	}
	return codes, nil
}

func (c *YuscomPublicCatalog) get(ctx context.Context, endpoint string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PulsaKilat-CatalogSync/1.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mengambil katalog Yuscom: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		resp.Body.Close()
		return nil, fmt.Errorf("mengambil katalog Yuscom: HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func findYuscomCatalogFrame(r io.Reader, pageURL string) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("membaca halaman katalog Yuscom: %w", err)
	}
	var src string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if src != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "iframe" {
			for _, attr := range node.Attr {
				if attr.Key == "src" && strings.Contains(strings.ToLower(attr.Val), "harga") {
					src = strings.TrimSpace(attr.Val)
					return
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if src == "" {
		return "", fmt.Errorf("iframe daftar harga Yuscom tidak ditemukan")
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(src)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(ref).String(), nil
}

func parseOpenYuscomProductCodes(r io.Reader) (map[string]struct{}, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("membaca tabel harga Yuscom: %w", err)
	}
	codes := make(map[string]struct{})
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "tr" {
			cells := directTableCells(node)
			if len(cells) >= 4 && strings.EqualFold(cells[3], "Open") {
				if code := strings.ToUpper(strings.TrimSpace(cells[0])); code != "" {
					codes[code] = struct{}{}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return codes, nil
}

func directTableCells(row *html.Node) []string {
	var cells []string
	for node := row.FirstChild; node != nil; node = node.NextSibling {
		if node.Type != html.ElementNode || node.Data != "td" {
			continue
		}
		cells = append(cells, strings.TrimSpace(nodeText(node)))
	}
	return cells
}

func nodeText(node *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			b.WriteString(current.Data)
			b.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(strings.Fields(b.String()), " ")
}
