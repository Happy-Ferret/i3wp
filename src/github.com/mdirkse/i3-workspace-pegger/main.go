package i3_workspace_pegger

import (
	"bytes"
	"fmt"
	"github.com/mdirkse/i3ipc"
	"github.com/spf13/afero"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	fs       = afero.NewOsFs()
	ipc      *i3ipc.IPCSocket
	varRegex = regexp.MustCompile("^set\\s+\\$([\\w]+)\\s+(.+)$")
	wsRegex  = regexp.MustCompile("^workspace\\s+(\\S+)\\s+output\\s+(\\S+)$")
)

func main() {
	var err error
	if ipc, err = i3ipc.GetIPCSocket(); err != nil {
		bail(err)
	}

	conf, err := readI3Config(ipc)
	if err != nil {
		bail(err)
	}

	vars, configWs := parseI3ConfigVariablesAndWorkspaces(conf)

	fmt.Printf("Variables:\n%s", dumpMap(vars))
	fmt.Printf("Workspaces:\n%s", dumpMap(configWs))

	resolvedWs := map[string]string{}
	for k, v := range configWs {
		rk := resolveVariables(k, vars)
		if rk == "" {
			fmt.Printf("WARNING: could not resolve variables in workspace name [%s]\n", k)
			continue
		}
		rv := resolveVariables(v, vars)
		if rv == "" {
			fmt.Printf("WARNING: could not resolve variables in output name [%s]\n", v)
		}
		resolvedWs[rk] = rv
	}

	fmt.Printf("Resolved workspaces:\n%s", dumpMap(resolvedWs))

	//outputs, err := ipcsocket.GetOutputs()
	//if err != nil {
	//	fmt.Errorf("Eeep3!", err)
	//}
	//
	//for _, output := range outputs {
	//	if !output.Active {
	//		continue
	//	}
	//
	//	outputB, _ := json.MarshalIndent(output, "", "  ")
	//	fmt.Printf("%+v\n", string(outputB))
	//}
	//
	//workspaces, err := ipcsocket.GetWorkspaces()
	//if err != nil {
	//	fmt.Errorf("Eeep4!", err)
	//}
	//
	//for _, ws := range workspaces {
	//	wsB, _ := json.MarshalIndent(ws, "", "  ")
	//	fmt.Printf("%+v\n", string(wsB))
	//}
}

func readI3Config(ipc *i3ipc.IPCSocket) ([]byte, error) {
	v, err := ipc.GetVersion()
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to get the location of the i3 config file (err: %s)", err)
	}
	fmt.Printf("Config file should be at [%s]\n", v.Loaded_Config_File_Name)

	if exists, err := afero.Exists(fs, v.Loaded_Config_File_Name); !exists || err != nil {
		if !exists && err == nil {
			err = fmt.Errorf("i3 config not found at [%s]", v.Loaded_Config_File_Name)
		}
		return []byte{}, err
	}

	c, err := afero.ReadFile(fs, v.Loaded_Config_File_Name)
	if err != nil {
		return []byte{}, err
	}

	return c, err

}

func resolveVariables(input string, vars map[string]string) string {
	if !strings.HasPrefix(input, "$") {
		return input
	}

	r, ok := vars[input[1:]]
	if !ok {
		fmt.Printf("WARNING: could not resolve variable [%s]\n", input)
	}
	return strings.TrimSpace(r)
}

func parseI3ConfigVariablesAndWorkspaces(i3Config []byte) (vars map[string]string, ws map[string]string) {
	vars = map[string]string{}

	lines := strings.Split(string(i3Config), "\n")

	for _, l := range lines {
		found := varRegex.FindSubmatch([]byte(strings.TrimSpace(l)))
		if found != nil {
			vars[string(found[1])] = strings.Trim(string(found[2]), "\"")
		}
	}

	ws = map[string]string{}

	for _, l := range lines {
		found := wsRegex.FindSubmatch([]byte(strings.TrimSpace(l)))
		if found != nil {
			ws[string(found[1])] = strings.Trim(string(found[2]), "\"")
		}
	}

	return vars, ws
}

func dumpMap(m map[string]string) string {
	var buffer bytes.Buffer

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf("\t%s -> %s\n", k, m[k]))
	}

	return buffer.String()
}
func bail(err error) {
	fmt.Printf("An error occurred. Cannot continue. Error: %s\n", err)
	os.Exit(1)
}
