package node

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Action struct {
	Name        string   `yaml:"name" json:"name"`
	Params      []string `yaml:"params" json:"params"`
	Command     string   `yaml:"command" json:"command"`
	Description string   `yaml:"description" json:"description"`
}

func (a Action) BuildCommand(params map[string]string) (string, error) {
	t, err := template.New(a.Name).Parse(a.Command)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := t.Execute(&b, params); err != nil {
		return "", err
	}
	result := b.String()
	return result, nil
}

func (a Action) Run(params map[string]string) (string, error) {
	cmdStr, err := a.BuildCommand(params)
	if err != nil {
		return "", err
	}

	args := strings.Fields(cmdStr)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func loadActions(filePath string) (map[string]Action, error) {
	yfile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var actions map[string]Action
	err2 := yaml.Unmarshal(yfile, &actions)
	if err2 != nil {
		return nil, err2
	}

	for name, a := range actions {
		if a.Name == "" {
			a.Name = name
		}
	}

	return actions, nil
}
