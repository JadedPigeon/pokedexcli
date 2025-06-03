package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	_ "github.com/JadedPigeon/pokedexcli/internal/pokecache"
)

// ===== Structs =====
type cliCommand struct {
	name        string
	description string
	callback    func(c *config) error
}

type config struct {
	nextLocationAreaURL     *string
	previousLocationAreaURL *string
}

type locationAreas struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// ===== Global Variables =====
var commands map[string]cliCommand

// ===== Helper Functions =====
func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	parts := strings.Fields(lower)
	// We only want the first word
	return parts
}

func commandExit(_ *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *config, cmds map[string]cliCommand) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cmds {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func GetURL(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

func getLocationAreas(c *config, url string) error {
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	}

	data, err := GetURL(url)
	if err != nil {
		return fmt.Errorf("error fetching location area: %w", err)
	}

	locations := locationAreas{}
	err = json.Unmarshal(data, &locations)
	if err != nil {
		return fmt.Errorf("error unmarshalling location area data: %w", err)
	}

	// Update the config with next and previous URLs
	c.nextLocationAreaURL = locations.Next
	c.previousLocationAreaURL = locations.Previous

	// Below prints are for debugging purposes
	for _, loc := range locations.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func commandMap(c *config) error {
	var url string
	if c.nextLocationAreaURL != nil {
		url = *c.nextLocationAreaURL
	}
	return getLocationAreas(c, url)
}

func commandMapb(c *config) error {
	if c.previousLocationAreaURL == nil {
		fmt.Println("You're on the first page.")
		return nil
	}
	return getLocationAreas(c, *c.previousLocationAreaURL)
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
			callback: func(c *config) error {
				return commandHelp(c, commands)
			},
		},
		"map": {
			name:        "map",
			description: "Display the next 20 locations in the Pokedex",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous 20 locations in the Pokedex",
			callback:    commandMapb,
		},
	}
}

// ===== Main Function =====
func main() {
	fmt.Println("Welcome to the Pokedex CLI!")
	cfg := &config{}
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
		if len(cleanline) == 0 {
			// No input, show prompt again
			continue
		}
		input := cleanline[0]
		if command, exists := commands[input]; exists {
			err := command.callback(cfg)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}

		fmt.Println("Your command was:", input)
	}
}
