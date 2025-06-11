package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

// FilterFunc return true if object contains field selector
type FilterFunc func(object runtime.Object) bool

func MatchObjectsByFieldSelector(data []byte, selector string) bool {
	fieldSelector, err := fields.ParseSelector(selector)
	if err != nil {
		return false
	}
	for _, requirement := range fieldSelector.Requirements() {
		var negative bool
		// supports '=', '==' and '!='.(e.g. ?fieldSelector=key1=value1,key2=value2)
		// fields.ParseSelector(FieldSelector) has handled the case where the operator is '==' and converted it to '=',
		// so case selection.DoubleEquals can be ignored here.
		switch requirement.Operator {
		case selection.NotEquals:
			negative = true
		case selection.Equals:
			negative = false
		default:
			return false
		}
		key := requirement.Field
		value := requirement.Value

		result, err := getStringByKey(data, key)
		if err != nil {
			return false
		}
		if (negative && fmt.Sprintf("%v", result) != value) ||
			(!negative && fmt.Sprintf("%v", result) == value) {
			continue
		}
		return false
	}
	return true
}

func parsePath(path string) []string {
	re := regexp.MustCompile(`[^\.\[\]]+|\[.*?\]`)

	matches := re.FindAllString(path, -1)

	for i, match := range matches {
		matches[i] = strings.Trim(match, "[]")
	}

	return matches
}

func getStringByKey(data []byte, key string) (string, error) {
	//fields.ParseSelector()
	jsonPath := parsePath(key)
	return jsonparser.GetString(data, jsonPath...)
}
