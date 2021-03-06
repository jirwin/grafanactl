package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/retzkek/grafanactl/gapi"
)

var helpCmd = &Command{
	Name:    "help",
	Usage:   "[COMMAND]",
	Summary: "Print command usage and options.",
	Help: `grafanactl uses the Grafana API to manage dashboards.
General Usage:
	grafanactl [OPTIONS] COMMAND [COMMAND OPTIONS]`,
}

func helpFunc(client *gapi.Client, cmd *Command, args []string) error {
	fmt.Printf("grafanactl v%s (%s/%s)\n\n", VERSION, REF, BUILD)
	switch {
	case cmd == nil, cmd == helpCmd && len(args) != 1:
		// no comand, just help command, or extra junk
		return printGeneralHelp()
	case cmd == helpCmd:
		// help command with subcommand
		if subcmd := findCommand(args[0]); subcmd != nil {
			return printCommandHelp(subcmd)
		} else {
			// unknown subcommand
			log.Error("unknown subcommand")
			return printGeneralHelp()
		}
	}
	// shouldn't get here
	log.Error("unhandled case in helpFunc switch")
	return printGeneralHelp()
}

func printGeneralHelp() error {
	tmpl, err := template.New("appHelp").Funcs(funcMap).Parse(generalHelpTemplate)
	if err != nil {
		return err
	}
	err = tmpl.Execute(os.Stdout, commands)
	if err != nil {
		return err
	}
	return nil
}

func printCommandHelp(cmd *Command) error {
	tmpl, err := template.New("comandHelp").Funcs(funcMap).Parse(commandHelpTemplate)
	if err != nil {
		return err
	}
	err = tmpl.Execute(os.Stdout, cmd)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	helpCmd.Function = helpFunc
}

var funcMap = template.FuncMap{
	"appFlags": func() string {
		var buf bytes.Buffer
		flag.VisitAll(func(f *flag.Flag) {
			if err := helpTemplates.ExecuteTemplate(&buf, "optionHelp", f); err != nil {
				panic(err)
			}
		})
		return buf.String()
	},
	"cmdFlags": func(cmdName string) string {
		var buf bytes.Buffer
		if cmd := findCommand(cmdName); cmd != nil {
			cmd.Flag.VisitAll(func(f *flag.Flag) {
				if err := helpTemplates.ExecuteTemplate(&buf, "optionHelp", f); err != nil {
					panic(err)
				}
			})
		}
		return buf.String()
	},
}

func renderTemplate(tmpl *template.Template, name string, data interface{}) string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		panic(err)
	}
	return buf.String()
}

var helpTemplates template.Template

func init() {
	helpTemplates.New("optionHelp").Parse(optionHelpTemplate)
}

var optionHelpTemplate = `
	-{{.Name}}=[{{.DefValue}}]
		{{.Usage}}`

var generalHelpTemplate = `
SYNOPSIS
	grafanactl is a backup/restore utility for Grafana dashboards.

USAGE
	grafanactl [OPTIONS] COMMAND [COMMAND OPTIONS]

OPTIONS
{{appFlags}}

COMMANDS
{{range .}}
	{{.Name}} {{.Usage}}
		{{.Summary}}
{{end}}
`

var commandHelpTemplate = `
{{.Name}}

SYNOPSIS
	{{.Summary}}

USAGE
	{{.Name}} {{.Usage}}

DESCRIPTION
	{{.Help}}

OPTIONS
{{cmdFlags .Name}}
`
