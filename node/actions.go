package node

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bofrim/gorch/hook"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

const StreamTeardownDelay = 250 * time.Millisecond

type Action struct {
	Name        string   `yaml:"name" json:"name"`
	Params      []string `yaml:"params" json:"params"`
	Commands    []string `yaml:"commands" json:"commands"`
	Description string   `yaml:"description" json:"description"`
}

type AdHocAction struct {
	ActionDef Action `yaml:"action" json:"action"`
}

func (a Action) BuildCommands(params any) ([]string, error) {
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

func (a Action) Run(params any) ([]string, error) {
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

func (a Action) RunStreamed(streamDest string, params any, logger *slog.Logger) error {
	commands, err := a.BuildCommands(params)
	if err != nil {
		logger.Error(
			"Failed to build action command.", err,
			slog.String("action", a.Name),
			slog.Any("params", params),
		)
		return err
	}

	hc := hook.NewHookClient(streamDest)
	go hc.Start()
	defer hc.Stop()

	for _, c := range commands {
		args := strings.Fields(c)
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.Output()
		if err != nil {
			logger.Error("Failed to get output for action command.", err,
				slog.String("action", a.Name),
				slog.String("command", c),
				slog.Any("params", params),
			)
			return err
		}
		if err := hc.Send(out); err != nil {
			logger.Error("Failed to send output for action command.", err,
				slog.String("action", a.Name),
				slog.String("command", c),
				slog.String("client", hc.Address),
			)
			return err
		}
	}
	time.Sleep(StreamTeardownDelay)
	return nil
}

func loadActions(filePath string) (map[string]*Action, error) {
	yfile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var actions map[string]*Action
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
