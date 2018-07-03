package utils

import (
	"regexp"
)

//embed regexp.Regexp in a new type so we can extend it
type ExRegexp struct {
	*regexp.Regexp
}

//add a new method to our new regular expression type
func (r *ExRegexp) FindStringSubmatchMap(s string) map[string]interface{} {
	captures := make(map[string]interface{})

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		//Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}
