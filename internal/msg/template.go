package msg

import (
	"regexp"
	"strings"
	"text/template"
)

func templateFuncs(overrides template.FuncMap) template.FuncMap {
	funcs := template.FuncMap{
		"lower": strings.ToLower,
		"regexReplaceAll": func(pattern, repl, s string) (string, error) {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return "", err
			}
			return re.ReplaceAllString(s, repl), nil
		},
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"trim": strings.TrimSpace,
		"trimPrefix": func(prefix, s string) string {
			return strings.TrimPrefix(s, prefix)
		},
		"trimSuffix": func(suffix, s string) string {
			return strings.TrimSuffix(s, suffix)
		},
		"upper": strings.ToUpper,
	}
	for name, fn := range overrides {
		funcs[name] = fn
	}
	return funcs
}
