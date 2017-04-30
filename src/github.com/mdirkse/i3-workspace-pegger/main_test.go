package i3_workspace_pegger

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveVariable(t *testing.T) {
	vars := map[string]string{
		"foo":   "bar",
		"hubba": "bubba",
	}

	testValues := map[string]string{
		"$foo":         "bar",
		"$hubba":       "bubba",
		"novar":        "novar",
		"$nonexistant": "",
	}

	for tv, rv := range testValues {
		assert.Equal(t, rv, resolveVariables(tv, vars), "Variable not resolved correctly!")
	}
}

func TestParseVariablesAndWorkspacesConfig(t *testing.T) {
	expectedVars := map[string]string{
		"foo":   "bar",
		"hubba": "bubba",
		"chit":  "chat",
		"lorem": "ipsum",
	}

	expectedWs := map[string]string{
		"$foo":  "$hubba",
		"$chit": "$lorem",
	}

	v, w := parseI3ConfigVariablesAndWorkspaces(getTestConfig())
	assert.Equal(t, expectedVars, v, "Wrong variables returned!")
	assert.Equal(t, expectedWs, w, "Wrong workspaces returned!")
}

func initTestEnvironment(i3ConfigLocation string) {
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, i3ConfigLocation, getTestConfig(), 0644)
}

func getTestConfig() []byte {
	return []byte(`
set $foo   bar
 set  $hubba bubba
set $chit  "chat"
set $lorem "ipsum"

workspace  $foo  output $hubba
 workspace $chit output  $lorem
`)
}
