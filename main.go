package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ===== Structs =====
type cliCommand struct {
	name        string
	description string
	callback    func() error
}

// ===== Global Variables =====
var commands map[string]cliCommand

// ===== Helper Functions =====
func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	parts := strings.Fields(lower)

	return parts
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cmds map[string]cliCommand) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cmds {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

// ===== Initialize =====
func init() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Show available commands",
			callback: func() error {
				return commandHelp(commands)
			},
		},
	}
}

// ===== Main Function =====
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanned := scanner.Scan()
		if !scanned {
			break
		}

		scanner := scanner.Text()
		cleanline := cleanInput(scanner)
		// We only care about the first word
		input := cleanline[0]
		if len(cleanline) == 0 {
			// No input, show prompt again
			continue
		}
		if command, exists := commands[input]; exists {
			command.callback()
		}

		fmt.Println("Your command was:", input)
	}
}
