package ui

import "net/url"

// resolveRelative превращает "/s.js" + "https://example.com/a" → "https://example.com/s.js".
func resolveRelative(baseURL, rel string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(rel)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(r).String(), nil
}