package utils

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type flagHost struct {
	str         string
	str2        string
	str3        string
	boolean     bool
	integer     int
	float       float64
	strArr      types.StringArray
	strMap      types.StringMap
	plainArr    []string
	plainMap    map[string]string
	unsupported struct{}
}

func TestConvertFlagsRegistersKnownTypes(t *testing.T) {
	host := &flagHost{str: "default", integer: 3, float: 1.5, plainArr: []string{"a"}, plainMap: map[string]string{"k": "v"}}
	flags := []types.Flag{
		{Name: "str", P: &host.str, V: host.str},
		{Name: "bool", P: &host.boolean, V: host.boolean},
		{Name: "int", P: &host.integer, V: host.integer},
		{Name: "float", P: &host.float, V: host.float},
		{Name: "str-arr", P: &host.strArr, V: host.strArr},
		{Name: "str-map", P: &host.strMap, V: host.strMap},
		{Name: "plain-arr", P: &host.plainArr, V: host.plainArr},
		{Name: "plain-map", P: &host.plainMap, V: host.plainMap},
		{Name: "unsupported", P: &host.unsupported, V: host.unsupported},
		{Name: "required", P: &host.str, V: host.str, Required: true},
	}

	cmd := &cobra.Command{Use: "test"}
	set := ConvertFlags(cmd, flags)

	for _, name := range []string{"str", "bool", "int", "float", "str-arr", "str-map", "plain-arr", "plain-map", "required"} {
		require.NotNil(t, set.Lookup(name), "flag %s must be registered", name)
	}
	assert.Nil(t, set.Lookup("unsupported"), "unknown value types are skipped")

	assert.Equal(t, "default", set.Lookup("str").DefValue)
	assert.Equal(t, "3", set.Lookup("int").DefValue)
	assert.Equal(t, "1.5", set.Lookup("float").DefValue)

	required := set.Lookup("required")
	require.NotNil(t, required.Annotations[cobra.BashCompOneRequiredFlag])
	assert.Equal(t, "true", required.Annotations[cobra.BashCompOneRequiredFlag][0])
}

func TestConvertFlagsShortHandAndEnvVar(t *testing.T) {
	host := &flagHost{}
	flags := []types.Flag{
		{Name: "str", P: &host.str, V: host.str, ShortHand: "s", EnvVar: "AUTOK3S_TEST_ENV"},
	}

	cmd := &cobra.Command{Use: "test"}
	set := ConvertFlags(cmd, flags)

	require.NotNil(t, set.ShorthandLookup("s"))
	require.NotNil(t, set.Lookup("str"))

	annotation := set.Lookup("str").Annotations[BashCompEnvVarFlag]
	require.NotNil(t, annotation)
	assert.Equal(t, "AUTOK3S_TEST_ENV", annotation[0])
}

func TestConvertFlagsIsIdempotent(t *testing.T) {
	host := &flagHost{}
	flags := []types.Flag{{Name: "str", P: &host.str, V: host.str}}

	cmd := &cobra.Command{Use: "test"}
	ConvertFlags(cmd, flags)
	ConvertFlags(cmd, flags)
}

func TestValidateRequiredFlags(t *testing.T) {
	host := &flagHost{}

	cmd := &cobra.Command{Use: "test"}
	set := ConvertFlags(cmd, []types.Flag{
		{Name: "with-default", P: &host.str, V: "preset", Required: true},
		{Name: "with-shorthand", P: &host.str2, V: "preset", ShortHand: "w", Required: true},
		{Name: "empty", P: &host.str3, V: "", Required: true},
	})

	// a required flag with a non-empty default passes the cobra check
	// without being set explicitly.
	ValidateRequiredFlags(set)

	assert.True(t, set.Lookup("with-default").Changed)
	assert.True(t, set.Lookup("with-shorthand").Changed)
	assert.False(t, set.Lookup("empty").Changed, "empty defaults stay unchanged")
}

func TestStringArrayValue(t *testing.T) {
	var arr types.StringArray
	v := newStringArrayValue([]string{"initial"}, &arr)
	assert.Equal(t, types.StringArray{"initial"}, arr)

	// the first Set replaces the initial value, later Sets append.
	require.NoError(t, v.Set("first"))
	require.NoError(t, v.Set("second"))
	assert.Equal(t, types.StringArray{"first", "second"}, arr)

	require.NoError(t, v.Replace([]string{"a", "b"}))
	assert.Equal(t, types.StringArray{"a", "b"}, arr)
	assert.Equal(t, []string{"a", "b"}, v.GetSlice())
	assert.Equal(t, "[a,b]", v.String())
	assert.Equal(t, "stringArray", v.Type())
}

func TestStringMapValue(t *testing.T) {
	var m types.StringMap
	v := newStringMapValue(map[string]string{}, &m)

	require.NoError(t, v.Set("a=1"))
	assert.Equal(t, types.StringMap{"a": "1"}, m)

	// a second Set merges instead of replacing.
	require.NoError(t, v.Set("b=2,c=3"))
	assert.Equal(t, types.StringMap{"a": "1", "b": "2", "c": "3"}, m)

	// values without a key=value shape are rejected.
	require.Error(t, v.Set("novalue"))

	assert.Equal(t, "stringMap", v.Type())
	assert.Contains(t, v.String(), "a=1")
}

func TestStringMapValueQuoted(t *testing.T) {
	var m types.StringMap
	v := newStringMapValue(map[string]string{}, &m)

	// a quoted pair is unquoted before parsing.
	require.NoError(t, v.Set(`"a=1"`))
	assert.Equal(t, types.StringMap{"a": "1"}, m)
}

func TestWriteAsCSV(t *testing.T) {
	out, err := writeAsCSV([]string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, "a,b", out)
}

func TestValidateRequiredFlagsNoopWithoutRequired(t *testing.T) {
	set := pflag.NewFlagSet("test", pflag.ContinueOnError)
	str := ""
	set.StringVar(&str, "plain", "", "")
	ValidateRequiredFlags(set)
	assert.False(t, set.Lookup("plain").Changed)
}
