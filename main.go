package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	parts := strings.Fields(lower)

	return parts
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanned := scanner.Scan()
		if !scanned {
			break
		}

		input := scanner.Text()
		cleanline := cleanInput(input)
		if len(cleanline) == 0 {
			// No input, show prompt again
			continue
		}

		fmt.Println("Your command was:", cleanline[0])
	}
}
