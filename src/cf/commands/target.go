package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"encoding/json"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"net/http"
)

type InfoResponse struct {
	ApiVersion            string `json:"api_version"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

func Target(c *cli.Context, ui term.UI) {
	if len(c.Args()) == 0 {
		showCurrentTarget(ui)
	} else {
		setNewTarget(c.Args()[0], ui)
	}

	return
}

func showCurrentTarget(ui term.UI) {
	config, err := configuration.Load()

	if err != nil {
		config = configuration.Default()
	}

	showConfiguration(config, ui)
}

func setNewTarget(target string, ui term.UI) {
	url := "https://" + target
	ui.Say("Setting target to %s...", term.Yellow(url))

	req, err := http.NewRequest("GET", url+"/v2/info", nil)

	if err != nil {
		ui.Failed("URL invalid.", err)
		return
	}

	client := api.NewClient()
	response, err := client.Do(req)

	if err != nil || response.StatusCode > 299 {
		ui.Failed("Target refused connection.", err)
		return
	}

	jsonBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		ui.Failed("Could not read response body.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = json.Unmarshal(jsonBytes, &serverResponse)

	if err != nil {
		ui.Failed("Invalid JSON response from server.", err)
		return
	}

	newConfiguration, err := saveTarget(url, serverResponse)

	if err != nil {
		ui.Failed("Error saving configuration", err)
		return
	}

	ui.Say(term.Green("OK"))
	showConfiguration(newConfiguration, ui)
}

func showConfiguration(config configuration.Configuration, ui term.UI) {
	ui.Say("CF instance: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	ui.Say("Logged out. Use '%s' to login.",
		term.Yellow("cf login USERNAME"))
}

func saveTarget(target string, info *InfoResponse) (config configuration.Configuration, err error) {
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}