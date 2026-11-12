package css

import "testing"

func TestParse(t *testing.T) {
	rules := ParseStylesheet(`
		html { color-scheme: light dark; background: light-dark(#eee,#222) }
		body { font: 16px/1.6 system-ui; max-width: 26em; margin: auto; padding: 25vh 2em 2em; text-align: center }
	`)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[1].Props["max-width"] != "26em" {
		t.Fatalf("bad max-width: %q", rules[1].Props["max-width"])
	}
}