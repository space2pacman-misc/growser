package css

import (
	"strings"
)

// Rule — одно CSS-правило: селектор + свойства.
type Rule struct {
	Selector string
	Props    map[string]string
}

// ParseStylesheet разбирает CSS-текст.
// Простой парсер: "sel1 { prop: val; ... } sel2 { ... }".
func ParseStylesheet(text string) []Rule {
	var rules []Rule
	text = stripComments(text)

	i := 0
	for i < len(text) {
		// селектор до "{"
		j := strings.IndexByte(text[i:], '{')
		if j < 0 {
			break
		}
		selector := strings.TrimSpace(text[i : i+j])
		i += j + 1

		// тело до "}"
		k := strings.IndexByte(text[i:], '}')
		if k < 0 {
			break
		}
		body := text[i : i+k]
		i += k + 1

		if selector == "" {
			continue
		}
		rules = append(rules, Rule{
			Selector: selector,
			Props:    parseDecls(body),
		})
	}
	return rules
}

func parseDecls(body string) map[string]string {
	props := make(map[string]string)
	for _, decl := range strings.Split(body, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		props[key] = val
	}
	return props
}

func stripComments(s string) string {
	for {
		a := strings.Index(s, "/*")
		if a < 0 {
			return s
		}
		b := strings.Index(s[a:], "*/")
		if b < 0 {
			return s[:a]
		}
		s = s[:a] + s[a+b+2:]
	}
}