package generator

import (
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/rancher/apiserver/pkg/types"
)

const (
	managementContextType = "mgmt"
)

var (
	underscoreRegexp = regexp.MustCompile(`([a-z])([A-Z])`)
)

func funcs() template.FuncMap {
	return template.FuncMap{
		"capitalize":          capitalize,
		"unCapitalize":        uncapitalize,
		"upper":               strings.ToUpper,
		"toLower":             strings.ToLower,
		"hasGet":              hasGet,
		"hasPost":             hasPost,
		"getCollectionOutput": getCollectionOutput,
		// "namespaced":          namespaced,
	}
}

func addUnderscore(input string) string {
	return strings.ToLower(underscoreRegexp.ReplaceAllString(input, `${1}_${2}`))
}

func hasGet(schema *types.APISchema) bool {
	return contains(schema.CollectionMethods, http.MethodGet)
}

// func namespaced(schema *types.APISchema) bool {
// 	return schema.Scope == types.NamespaceScope
// }

func hasPost(schema *types.APISchema) bool {
	return contains(schema.CollectionMethods, http.MethodPost)
}

func contains(list []string, needle string) bool {
	for _, i := range list {
		if i == needle {
			return true
		}
	}
	return false
}

func getCollectionOutput(output, codeName string) string {
	if output == "collection" {
		return codeName + "Collection"
	}
	return capitalize(output)
}

// // SyncOnlyChangedObjects check whether the CATTLE_SKIP_NO_CHANGE_UPDATE env var is
// // configured to skip the update handler for events on the management context
// // that do not contain a change to the object.
// func SyncOnlyChangedObjects() bool {
// 	skipNoChangeUpdate := os.Getenv("CATTLE_SYNC_ONLY_CHANGED_OBJECTS")
// 	if skipNoChangeUpdate == "" {
// 		return false
// 	}
// 	parts := strings.Split(skipNoChangeUpdate, ",")

// 	for _, part := range parts {
// 		if part == managementContextType {
// 			return true
// 		}
// 	}
// 	return false
// }

func capitalize(s string) string {
	if len(s) <= 1 {
		return strings.ToUpper(s)
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func uncapitalize(s string) string {
	if len(s) <= 1 {
		return strings.ToLower(s)
	}

	return strings.ToLower(s[:1]) + s[1:]
}
