package css

import "strings"

// Match проверяет, подходит ли селектор к элементу с тегом tag и классами classes.
func Match(selector, tag string, classes map[string]bool) bool {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return false
	}
	// пока без пробелов и запятых (потомки, группы) — просто отсекаем их
	if strings.ContainsAny(selector, " ,>+~") {
		// группы через запятую поддержим в цикле — здесь лишь один компонент
		selector = strings.Fields(selector)[0]
	}

	// #id
	if strings.HasPrefix(selector, "#") {
		id := selector[1:]
		return classes["#"+id] // см. classes ниже — храним id с # в рендере
	}

	// .class
	if strings.HasPrefix(selector, ".") {
		return classes[selector[1:]]
	}

	// tag.class
	if dot := strings.IndexByte(selector, '.'); dot > 0 {
		wantTag := selector[:dot]
		wantClass := selector[dot+1:]
		return wantTag == tag && classes[wantClass]
	}

	// tag
	return selector == tag
}

// ComputeProperties собирает свойства для элемента: от всех подходящих правил
// (порядок — как в stylesheet, последнее побеждает), плюс inline-стили.
func ComputeProperties(rules []Rule, tag string, classes map[string]bool, inline string) map[string]string {
	out := make(map[string]string)

	// Для каждого правила проверяем КАЖДЫЙ компонент через запятую
	for _, rule := range rules {
		for _, sel := range strings.Split(rule.Selector, ",") {
			if Match(sel, tag, classes) {
				for k, v := range rule.Props {
					out[k] = v
				}
				break
			}
		}
	}

	if inline != "" {
		for k, v := range parseDecls(inline) {
			out[k] = v
		}
	}
	return out
}