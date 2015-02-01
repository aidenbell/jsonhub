package ext_ci_match

import "strings"

/*
 * Case insensitive matcher. sval will be the part of the
 * MatchSpec referencing this extension. mval is the message value
 * of unknown type.
 */
func ExtCaseInsensitiveMatch(mval interface{}, sval map[string]interface{}) bool {
	specif, ok := sval["value"]
	if !ok {
		return false
	}

	specval, ok := specif.(string)
	if !ok {
		return false
	}

	switch mcast := mval.(type) {
	case string:
		if strings.ToLower(specval) == strings.ToLower(mcast) {
			return true
		}
	}
	return false
}
