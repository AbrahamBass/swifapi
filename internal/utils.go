package swiftapi

import (
	"path"
	"regexp"
	"strings"

	"github.com/AbrahamBass/swifapi/internal/types"
)

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}

type pathMatcher struct {
	original   string
	regex      *regexp.Regexp
	paramNames []string
}

func NewPathMatcher(original string, regex *regexp.Regexp, paramName []string) pathMatcher {
	return pathMatcher{
		original:   original,
		regex:      regex,
		paramNames: paramName,
	}
}

func (rp *pathMatcher) Match(path string) (bool, map[string]string) {
	matches := rp.regex.FindStringSubmatch(path)
	if matches == nil {
		return false, nil
	}

	params := make(map[string]string)
	for i, name := range rp.paramNames {
		params[name] = matches[i+1]
	}
	return true, params
}

func compilePattern(pattern string) pathMatcher {
	regexPattern := "^" + strings.ReplaceAll(pattern, "{", "(?P<") + "$"
	regexPattern = strings.ReplaceAll(regexPattern, "}", ">[^/]+)")

	re := regexp.MustCompile(regexPattern)
	paramNames := re.SubexpNames()[1:]

	return NewPathMatcher(pattern, re, paramNames)
}

func combineMiddlewares(global, group, route []types.Middleware) []types.Middleware {
	combined := make([]types.Middleware, 0, len(global)+len(group)+len(route))
	combined = append(combined, global...)
	combined = append(combined, group...)
	combined = append(combined, route...)
	return combined
}
