package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform/helper/hashcode"
)

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func isMap(i interface{}) bool {
	return (reflect.ValueOf(i).Kind() == reflect.Map)
}

func isSlice(i interface{}) bool {
	return (reflect.ValueOf(i).Kind() == reflect.Slice)
}

func quoteIfString(i interface{}) interface{} {
	// Handle <no value> zero value by converting it to an empty string
	if i == nil {
		return "\"\""
	}
	if reflect.ValueOf(i).Kind() == reflect.String {
		return fmt.Sprintf("\"%v\"", i)
	}
	return i
}

func normalizeResourceName(name string) string {
	r := strings.NewReplacer(".", "_", "*", "star", " ", "_")

	return r.Replace(name)
}

func hasLeadingDigit(identifier string) bool {
	_, err := strconv.ParseFloat(identifier[:1], 64)
	return err == nil
}

var templateFuncMap = template.FuncMap{
	"replace":               replace,
	"isMap":                 isMap,
	"isSlice":               isSlice,
	"quoteIfString":         quoteIfString,
	"trim":                  strings.TrimSpace,
	"normalizeResourceName": normalizeResourceName,
	"hasLeadingDigit":       hasLeadingDigit,
}

func hashMap(values map[string]string) int {
	var keys []string
	var buf bytes.Buffer

	for k := range values {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	buf.WriteString("{<")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteRune(':')
		buf.WriteString(values[k])
		buf.WriteRune(';')
	}
	buf.WriteString(">;};")

	return hashcode.String(buf.String())
}
