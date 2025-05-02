package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const deployScriptFile = "deploy.json"

type DeployScript struct {
	Commands []string `json:"commands"`
}

func FileDeployCheck() error {
	localPath, err := os.Getwd()
	if err != nil {
		return err
	}

	deployFile := filepath.Join(localPath, deployScriptFile)
	_, err = os.Stat(deployFile)
	if err != nil {
		msg := T("file-err-missing-deploy")
		return fmt.Errorf(msg)
	}

	return nil
}

func FileDeployRead() (*DeployScript, error) {
	localPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	deployFile := filepath.Join(localPath, deployScriptFile)
	deployData, err := os.ReadFile(deployFile)
	if err != nil {
		if os.IsNotExist(err) {
			msg := T("file-err-missing-deploy")
			return nil, fmt.Errorf(msg)
		}
		return nil, err
	}

	var deployScript DeployScript
	if err := json.Unmarshal(deployData, &deployScript); err != nil {
		return nil, err
	}

	return &deployScript, nil
}

func (ds *DeployScript) Save() error {
	content, err := json.MarshalIndent(*ds, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(deployScriptFile, content, 0644); err != nil {
		return err
	}

	return nil
}
