package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

// HarvestPlugin is the Plugin implementation
type HarvestPlugin struct {
}

// Result from apps json
type Result struct {
	Resources []struct {
		Name     string `json:"name"`
		GUID     string `json:"guid"`
		Metadata struct {
			Labels struct {
				KeepUntil string `json:"keep_until"`
			} `json:"labels"`
		} `json:"metadata"`
	} `json:"resources"`
}

// Run parses command line args and runs respective commands
func (p *HarvestPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	var err error

	appsURL := "/v3/apps"
	// for below: all apps you have auth to destroy (across orgs/spaces)
	// whatever query to get app guids with reap_on key

	// get appss
	// apps, _ := cliConnection.GetApps()
	appsJSONResult, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", appsURL)
	// fmt.Printf("%s\n\n", appsJSONResult)

	appsJSONString := strings.Join(appsJSONResult, "")

	var result Result
	if err := json.Unmarshal([]byte(appsJSONString), &result); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n\n", result)
	apps := result.Resources

	// apps = json.parse(appsJSONString, Result)

	// for each app with reap_on
	for _, a := range apps {
		// if value < current time
		var keepUntilTime int
		unset := (a.Metadata.Labels.KeepUntil == "")

		if !unset {
			keepUntilTime, err = strconv.Atoi(a.Metadata.Labels.KeepUntil)
			if err != nil {
				panic(err)
			}
		}
		harvestTime := (int64(keepUntilTime) > time.Now().Unix())

		fmt.Printf("DEBUG name:%s unset:%b harvestTime:%b\n", a.Name, unset, harvestTime)
		if unset || harvestTime {
			fmt.Print("curl ", fmt.Sprintf("/v3/apps/%s ", a.GUID), "-X ", "PUT ", "-d ", `{"state": "STOPPED"}`)

			cliConnection.CliCommandWithoutTerminalOutput(
				"curl", fmt.Sprintf("/v3/apps/%s", a.GUID), "-X", "PUT", "-d", `{"state": "STOPPED"}`)
		}

	}

}

// GetMetadata for plugin
func (p *HarvestPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CloudFoundryHarvestPlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "harvest",
				HelpText: "Manage your org/space/app usage harvest",

				UsageDetails: plugin.Usage{
					Usage: "harvest\n   cf harvest",
				},
			},
		},
	}
}

func main() {

	plugin.Start(new(HarvestPlugin))
}
