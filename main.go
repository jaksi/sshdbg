package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type command struct {
	aliases     []string
	description string
	usage       string
	action      func(args []string) error
}

var (
	exit  = errors.New("exit")
	usage = errors.New("usage")
)

var (
	terminal *term.Terminal
)

var commands = []command{
	{
		aliases:     []string{"exit", "quit"},
		description: "exit the program",
		usage:       "",
		action: func(args []string) error {
			if len(args) != 0 {
				return usage
			}
			return exit
		},
	},
}

func init() {
	commands = append(commands, command{
		aliases:     []string{"help", "h"},
		description: "list available commands",
		usage:       "",
		action: func(args []string) error {
			if len(args) != 0 {
				return usage
			}
			for _, cmd := range commands {
				fmt.Fprintf(terminal, "%s\n%s\nUsage: %s %s\n\n", strings.Join(cmd.aliases, "|"), cmd.description, cmd.aliases[0], cmd.usage)
			}
			return nil
		},
	})
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState) //nolint:errcheck
	terminal = term.NewTerminal(os.Stdin, "sshdbg> ")
	terminal.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key != '\t' {
			return line, pos, false
		}
		if pos != len(line) {
			return line, pos, false
		}
		for _, cmd := range commands {
			for _, alias := range cmd.aliases {
				if strings.HasPrefix(alias, line) {
					return alias, len(alias), true
				}
			}
		}
		return line, pos, false
	}
	for {
		line, err := terminal.ReadLine()
		if err != nil {
			break
		}
		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		var cmd *command
		for _, c := range commands {
			for _, a := range c.aliases {
				if a == args[0] {
					cmd = &c
					break
				}
			}
			if cmd != nil {
				break
			}
		}
		if cmd == nil {
			fmt.Fprintf(terminal, "Unknown command: %s\n", args[0])
			continue
		}
		if err := cmd.action(args[1:]); err != nil {
			if err == exit {
				break
			}
			if err == usage {
				fmt.Fprintf(terminal, "Usage: %s %s\n", args[0], cmd.usage)
				continue
			}
			fmt.Fprintf(terminal, "Error: %s\n", err)
		}
	}
}
