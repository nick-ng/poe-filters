package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	colourTokens map[string]string
)

func init() {
	colourJsonBytes, err := os.ReadFile(filepath.Join("utils", "colour-tokens.json"))

	if err != nil {
		fmt.Println("error loading colour-tokens.json", err)
		os.Exit(1)
	}

	err = json.Unmarshal(colourJsonBytes, &colourTokens)

	if err != nil {
		fmt.Println("invalid colour-tokens.json file", err)
		os.Exit(1)
	}

	fmt.Println(colourTokens["DefaultBackground"])
}

// @todo(nick-ng): replace armour group tokens
func ApplyAllTokens(rawFilter string) string {
	processedFilter := applyColourTokens(rawFilter)

	remainingTokensRe := regexp.MustCompile(`#!.+!#`)

	processedFilter = remainingTokensRe.ReplaceAllString(processedFilter, "")

	return processedFilter
}

func applyColourTokens(rawFilter string) string {
	processedFilter := rawFilter

	for key, value := range colourTokens {
		token := fmt.Sprintf("#!%s!#", key)
		processedFilter = strings.ReplaceAll(processedFilter, token, value)
	}

	return processedFilter
}
