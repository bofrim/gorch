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
	Commands    []string `yaml:"commands" json:"commands"`
	Description string   `yaml:"description" json:"description"`
}

func (a Action) BuildCommands(params map[string]string) ([]string, error) {

	commands := make([]string, len(a.Commands))
	for i, command := range a.Commands {

		t, err := template.New(a.Name).Parse(command)
		if err != nil {
			return nil, err
		}

		var b bytes.Buffer
		if err := t.Execute(&b, params); err != nil {
			return nil, err
		}
		commands[i] = b.String()
	}
	return commands, nil
}

func (a Action) Run(params map[string]string) ([]string, error) {
	commands, err := a.BuildCommands(params)
	if err != nil {
		return nil, err
	}

	results := make([]string, len(commands))

	for i, c := range commands {
		args := strings.Fields(c)
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		results[i] = string(out)

	}
	return results, nil
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
