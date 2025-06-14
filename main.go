package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JadedPigeon/pokedexcli/internal/pokecache"
)

// ===== Structs =====
type cliCommand struct {
	name        string
	description string
	callback    func(c *config, args []string) error
}

type config struct {
	nextLocationAreaURL     *string
	previousLocationAreaURL *string
	cache                   *pokecache.Cache
	pokedex                 map[string]PokemonInfo
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

type AreaEncounter struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type PokemonInfo struct {
	Name   string           `json:"name"`
	Height int              `json:"height"`
	Weight int              `json:"weight"`
	Exp    int              `json:"base_experience"`
	Stats  []StatInfo       `json:"stats"`
	Types  []TypeInfoParent `json:"types"`
}

type StatInfo struct {
	BaseStat int           `json:"base_stat"`
	Stat     StatNameField `json:"stat"`
}

type StatNameField struct {
	Name string `json:"name"`
}

type TypeInfoParent struct {
	Type TypeNameField `json:"type"`
}

type TypeNameField struct {
	Name string `json:"name"`
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

func commandExit(_ *config, _ []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *config, cmds map[string]cliCommand, _ []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cmds {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func GetURL(url string, cache *pokecache.Cache) ([]byte, error) {
	if data, ok := cache.Get(url); ok {
		//fmt.Println("[Cache hit]", url)
		return data, nil
	}

	// Simulate long network delay
	// fmt.Println("[Cache miss] Waiting 2 seconds before fetching", url)
	// time.Sleep(2 * time.Second)

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
	// fmt.Println("[Cache miss] adding to cache", url)
	cache.Add(url, body)
	return body, nil
}

func getLocationAreas(c *config, url string) error {
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"
	}

	data, err := GetURL(url, c.cache)
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
	// Print the next and previous URLs for debugging
	// if c.nextLocationAreaURL != nil {
	// 	fmt.Println("Next URL (value):", *c.nextLocationAreaURL)
	// }
	// if c.previousLocationAreaURL != nil {
	// 	fmt.Println("Previous URL (value):", *c.previousLocationAreaURL)
	// }

	for id, loc := range locations.Results {
		fmt.Printf("Location %d: %s\n", id+1, loc.Name)
	}
	return nil
}

func commandMap(c *config, _ []string) error {
	var url string
	if c.nextLocationAreaURL != nil {
		url = *c.nextLocationAreaURL
	}
	return getLocationAreas(c, url)
}

func commandMapb(c *config, _ []string) error {
	if c.previousLocationAreaURL == nil {
		fmt.Println("You're on the first page.")
		return nil
	}
	return getLocationAreas(c, *c.previousLocationAreaURL)
}

func commandExplore(c *config, id []string) error {
	if len(id) == 0 || id[0] == "" {
		return fmt.Errorf("please provide a location area ID or name")
	}
	url := "https://pokeapi.co/api/v2/location-area/" + id[0]
	data, err := GetURL(url, c.cache)
	if err != nil {
		return fmt.Errorf("error fetching location area %s: %w", id[0], err)
	}

	area := AreaEncounter{}
	err = json.Unmarshal(data, &area)
	if err != nil {
		return fmt.Errorf("error unmarshalling area encounter data: %w", err)
	}
	fmt.Printf("Exploring location area %s:\n", id[0])
	for _, encounter := range area.PokemonEncounters {
		fmt.Printf("  - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(c *config, name []string) error {
	if len(name) == 0 || name[0] == "" {
		return fmt.Errorf("please provide a pokemone's name")
	}
	url := "https://pokeapi.co/api/v2/pokemon/" + name[0]
	data, err := GetURL(url, c.cache)
	if err != nil {
		return fmt.Errorf("error fetching pokemon data, please check spelling: %w", err)
	}

	pokemon := PokemonInfo{}
	err = json.Unmarshal(data, &pokemon)
	if err != nil {
		return fmt.Errorf("error unmarshalling pokemone data: %w", err)
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	chance := float32((400 - pokemon.Exp)) / 4
	randomchance := rand.Float32() * 100
	if chance > randomchance {
		key := strings.ToLower(pokemon.Name)
		c.pokedex[key] = pokemon
		fmt.Printf("%s was caught!\n", pokemon.Name)
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}
	return nil
}

func commandInspect(c *config, name []string) error {
	if len(name) == 0 || name[0] == "" {
		return fmt.Errorf("please provide a pokemone's name")
	}
	key := strings.ToLower(name[0])
	pokemon, ok := c.pokedex[key]
	if !ok {
		fmt.Println("You have not caught that Pokémon.")
		return nil
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(c *config, _ []string) error {
	if len(c.pokedex) == 0 {
		fmt.Println("No pokemon have been caught yet")
	}
	fmt.Println("Your Pokedex:")
	for _, name := range c.pokedex {
		fmt.Printf(" - %s\n", name.Name)
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
			callback: func(c *config, _ []string) error {
				return commandHelp(c, commands, nil)
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
		"explore": {
			name:        "explore",
			description: "Explore the location: Usage: explore <location_id>",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a pokemone: Usage catch <pokemon>",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a pokemon you have caught",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists all the pokemone you have caught",
			callback:    commandPokedex,
		},
	}
}

// ===== Main Function =====
func main() {
	fmt.Println("Welcome to the Pokedex CLI!")
	cfg := &config{
		cache:   pokecache.NewCache(60 * time.Second),
		pokedex: make(map[string]PokemonInfo),
	}
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
		args := cleanline[1:]
		if command, exists := commands[input]; exists {
			err := command.callback(cfg, args)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}

		fmt.Println("Your command was:", input)
	}
}
