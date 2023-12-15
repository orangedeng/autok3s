/*
Package generator contains templates and functions to generate clients, controllers, and lifecycles. The Generate functions
are exported and intended to be used by other libraries.
*/
package generator

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/schemas"
	"golang.org/x/tools/imports"
	"k8s.io/gengo/args"
)

var (
	blackListTypes = map[string]bool{
		"provider":   true,
		"schema":     true,
		"resource":   true,
		"collection": true,
	}
)

func GenerateClient(schemas *types.APISchemas, privateTypes map[string]bool, outputDir, cattleOutputPackage string) error {
	baseDir := args.DefaultSourceTree()
	cattleDir := path.Join(outputDir, cattleOutputPackage)

	if err := prepareDirs(cattleDir); err != nil {
		return err
	}

	var cattleClientTypes []*types.APISchema
	for _, schema := range schemas.Schemas {
		if blackListTypes[schema.ID] {
			continue
		}

		if err := generateType(cattleDir, schema, schemas); err != nil {
			return err
		}
		if _, privateType := privateTypes[schema.ID]; !privateType {
			cattleClientTypes = append(cattleClientTypes, schema)
		}
	}

	if err := generateClient(cattleDir, cattleClientTypes); err != nil {
		return err
	}

	return Gofmt(baseDir, filepath.Join(outputDir, cattleOutputPackage))
}

func generateType(outputDir string, schema *types.APISchema, schemas *types.APISchemas) error {
	filePath := strings.ToLower("zz_generated_" + addUnderscore(schema.ID) + ".go")
	output, err := os.Create(path.Join(outputDir, filePath))
	if err != nil {
		return err
	}
	defer output.Close()

	typeTemplate, err := template.New("type.template").
		Funcs(funcs()).
		Parse(strings.Replace(typeTemplate, "%BACK%", "`", -1))
	if err != nil {
		return err
	}

	return typeTemplate.Execute(output, map[string]interface{}{
		"schema":            schema,
		"structFields":      getTypeMap(schema, schemas),
		"resourceActions":   getResourceActions(schema, schemas),
		"collectionActions": getCollectionActions(schema, schemas),
	})
}

func generateClient(outputDir string, schemas []*types.APISchema) error {
	template, err := template.New("client.template").
		Funcs(funcs()).
		Parse(clientTemplate)
	if err != nil {
		return err
	}

	output, err := os.Create(path.Join(outputDir, "zz_generated_client.go"))
	if err != nil {
		return err
	}
	defer output.Close()

	return template.Execute(output, map[string]interface{}{
		"schemas": schemas,
	})
}

type fieldInfo struct {
	Name string
	Type string
}

func getGoType(field schemas.Field, schema *schemas.Schema, schemas *schemas.Schemas) string {
	return getTypeString(field.Nullable, field.Type, false, schema, schemas)
}

func getTypeString(nullable bool, typeName string, pointer bool, schema *schemas.Schema, schemas *schemas.Schemas) string {
	switch {
	case pointer:
		return "*" + getTypeString(nullable, typeName, false, schema, schemas)
	case strings.HasPrefix(typeName, "reference["):
		return "string"
	case strings.HasPrefix(typeName, "map["):
		return "map[string]" + getTypeString(false, typeName[len("map["):len(typeName)-1], false, schema, schemas)
	case strings.HasPrefix(typeName, "array["):
		return "[]" + getTypeString(false, typeName[len("array["):len(typeName)-1], false, schema, schemas)
	}

	name := ""

	switch typeName {
	case "base64":
		return "string"
	case "json":
		return "interface{}"
	case "boolean":
		name = "bool"
	case "float":
		name = "float64"
	case "int":
		name = "int64"
	case "multiline":
		return "string"
	case "masked":
		return "string"
	case "password":
		return "string"
	case "date":
		return "string"
	case "string":
		return "string"
	case "enum":
		return "string"
	case "intOrString":
		return "intstr.IntOrString"
	case "dnsLabel":
		return "string"
	case "dnsLabelRestricted":
		return "string"
	case "hostname":
		return "string"
	default:
		if schema != nil && schemas != nil {
			otherSchema := schemas.Schema(typeName)
			if otherSchema != nil {
				name = otherSchema.CodeName
			}
		}

		if name == "" {
			name = capitalize(typeName)
		}
	}

	if nullable {
		return "*" + name
	}

	return name
}

func getTypeMap(schema *types.APISchema, schemas *types.APISchemas) map[string]fieldInfo {
	result := map[string]fieldInfo{}
	for name, field := range schema.ResourceFields {
		if strings.EqualFold(name, "id") && hasGet(schema) {
			continue
		}
		result[field.CodeName] = fieldInfo{
			Name: name,
			Type: getGoType(field, schema.InternalSchema, schemas.InternalSchemas),
		}
	}
	return result
}

func getResourceActions(schema *types.APISchema, ss *types.APISchemas) map[string]schemas.Action {
	result := map[string]schemas.Action{}
	for name, action := range schema.ResourceActions {
		if action.Output != "" {
			if ss.InternalSchemas.Schema(action.Output) != nil {
				result[name] = action
			}
		} else {
			result[name] = action
		}
	}
	return result
}

func getCollectionActions(schema *types.APISchema, ss *types.APISchemas) map[string]schemas.Action {
	result := map[string]schemas.Action{}
	for name, action := range schema.CollectionActions {
		if action.Output != "" {
			output := action.Output
			if action.Output == "collection" {
				output = strings.ToLower(schema.CodeName)
			}
			if ss.InternalSchemas.Schema(output) != nil {
				result[name] = action
			}
		} else {
			result[name] = action
		}
	}
	return result
}

func prepareDirs(dirs ...string) error {
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), "zz_generated") {
				if err := os.Remove(path.Join(dir, file.Name())); err != nil {
					return errors.Wrapf(err, "failed to delete %s", path.Join(dir, file.Name()))
				}
			}
		}
	}

	return nil
}

func Gofmt(workDir, pkg string) error {
	return filepath.Walk(filepath.Join(workDir, pkg), func(path string, info os.FileInfo, _ error) error {
		println(path)
		if info.IsDir() {

			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := imports.Process(path, content, &imports.Options{
			Fragment:   true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   8,
			FormatOnly: true,
		})
		if err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(formatted)
		return err
	})
}
