// Package ext_ci_match provides case insensitive matching for subscription
// patterns. The value attribute contains the string to test against the string
// value in the "target" attribute of the message (or whatever attribute)
// you are matching against.
//
// 	"target": {
//		"__match__": "case-insensitive",
//		"value": "ValueToTest"
// 	}
//
// If the target value in the message is not a string, a false match is returned
package ext_ci_match

import "strings"

// ExtCaseInsensitiveMatch is a pattern extension to test if a message value
// (mval) has case insensitive equality with the subscription value (sval)
// the subscription value is the whole match object in the format
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
