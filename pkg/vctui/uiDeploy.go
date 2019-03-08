package vctui

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/plunder-app/plunder/pkg/server"
	"github.com/rivo/tview"
)

func deployOnVM(address, hostname string) {
	uiBugFix()

	app := tview.NewApplication()

	form := tview.NewForm()
	form.AddInputField("Plunder Address", "http://localhost", 40, nil, nil).
		AddInputField("MAC Address", address, 18, nil, nil).
		AddInputField("Hostname", hostname, 40, nil, nil).
		AddInputField("IP Address", "", 18, nil, nil).
		AddDropDown("Deployment Type", deplopyTypes, 0, nil).
		AddButton("Save Settings", func() { app.Stop() })

	form.SetBorder(true).
		SetTitle("Update deployment").
		SetTitleAlign(tview.AlignCenter).
		SetRect(5, 1, 60, 23)

	if err := app.SetRoot(form, false).Run(); err != nil {
		panic(err)
	}
	var deployMac, deployHost, deployIP, deployType string

	plunderURL := form.GetFormItemByLabel("Plunder Address").(*tview.InputField).GetText()
	deployMac = form.GetFormItemByLabel("MAC Address").(*tview.InputField).GetText()
	deployHost = form.GetFormItemByLabel("Hostname").(*tview.InputField).GetText()
	deployIP = form.GetFormItemByLabel("IP Address").(*tview.InputField).GetText()
	_, deployType = form.GetFormItemByLabel("Deployment Type").(*tview.DropDown).GetCurrentOption()

	currentConfig, err := getConfig(plunderURL)

	if err != nil {
		errorUI(err)
		return
	}
	// Check for existing deployment
	var updatedExisting bool
	for i := range currentConfig.Deployments {
		if currentConfig.Deployments[i].MAC == deployMac {
			currentConfig.Deployments[i].Deployment = deployType
			currentConfig.Deployments[i].Config.IPAddress = deployIP
			currentConfig.Deployments[i].Config.ServerName = deployHost
			updatedExisting = true
		}
	}

	// If we've not updated the existing then it's a new entry
	if updatedExisting == false {
		newDeployment := server.DeploymentConfigurations{
			MAC:        deployMac,
			Deployment: deployType,
			Config: server.HostConfig{
				IPAddress:  deployIP,
				ServerName: deployHost,
			},
		}
		currentConfig.Deployments = append(currentConfig.Deployments, newDeployment)
	}

	// Update the deployment server
	err = postConfig(plunderURL, currentConfig)
	if err != nil {
		errorUI(err)
		return
	}
}

func getConfig(plunderURL string) (*server.DeploymentConfigurationFile, error) {
	u, err := url.Parse(plunderURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/deployment"

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	var config server.DeploymentConfigurationFile
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func postConfig(plunderURL string, config *server.DeploymentConfigurationFile) error {
	u, err := url.Parse(plunderURL)
	if err != nil {
		return err
	}
	u.Path = "/deployment"

	jsonValue, _ := json.Marshal(config)

	_, err = http.Post(u.String(), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	return nil
}
