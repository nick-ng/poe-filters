package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MY_FILTERS_PATH string = "my-filters"
const BASE_FILTERS_PATH string = "base-filters"
const THIRD_PARTY_FILTERS_PATH string = "third-party-filters"
const OUTPUT_FILTERS_PATH string = "output-filters"

func main() {
	err := os.Mkdir(MY_FILTERS_PATH, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating input filters directory", err)
		return
	}

	err = os.Mkdir(OUTPUT_FILTERS_PATH, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating output filters directory", err)
		return
	}

	err = os.Mkdir(BASE_FILTERS_PATH, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating base filters directory", err)
		return
	}

	err = os.Mkdir(THIRD_PARTY_FILTERS_PATH, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating third party filters directory", err)
		return
	}

	path1 := filepath.Join(MY_FILTERS_PATH)

	dat1, err := os.ReadDir(path1)

	if err != nil {
		fmt.Println("Error reading file", err)
	}

	fmt.Printf("Found %d filters\n", len(dat1))

	for i, dir := range dat1 {
		if err != nil {
			continue
		}

		filterName := dir.Name()

		fmt.Printf("%d: %s...", i+1, filterName)

		path := filepath.Join(MY_FILTERS_PATH, filterName)

		filter, errList := processFilter(path)

		filterData := []byte(filter)

		path = filepath.Join(OUTPUT_FILTERS_PATH, filterName)
		err := os.WriteFile(path, filterData, 0666)

		if err != nil {
			fmt.Println("\nError writing filter", filterName, err)
			continue
		}

		if len(errList) > 0 {
			fmt.Printf(" done with %d errors\n", len(errList))
			continue
		} else {
			fmt.Println(" done")
		}
	}
}

func processFilter(filterPath string) (string, []error) {
	var errorList []error

	filterData, err := os.ReadFile(filterPath)

	if err != nil {
		return "", append(errorList, err)
	}

	rawLines := strings.Split(string(filterData), "\n")

	processedLines := []string{}

	var currentCommand string
	options := make(map[string]string)

	for _, rawLine := range rawLines {
		switch currentCommand {
		case "import":
			{
				if strings.HasPrefix(rawLine, "#!") {
					// @todo(nick-ng): handle delete options in import commands
					processedLines = append(processedLines, rawLine)
				} else if strings.HasPrefix(rawLine, "#") {
					processedLines = append(processedLines, rawLine)
				} else {
					// it's a non-comment line so import the filter here then write the line
					_, present := options["file"]
					if present {
						importedFilter, err := importBaseFilter(options["file"])
						if err != nil {
							errorList = append(errorList, err)
							processedLines = append(processedLines, fmt.Sprintf("# Error: Couldn't import %s", options["file"]))
							processedLines = append(processedLines, fmt.Sprintf("#  %s\n", err))
						} else {
							processedLines = append(processedLines, importedFilter)
							processedLines = append(processedLines, fmt.Sprintf("# End of %s\n", options["file"]))
						}
						processedLines = append(processedLines, rawLine)
					} else {
						processedLines = append(processedLines, "# Error: No file specified.\n")
						processedLines = append(processedLines, rawLine)
					}

					currentCommand = ""
					clear(options)
				}
				break
			}
		default:
			{
				if strings.HasPrefix(rawLine, "#!") {
					// #! <command> <argument1> <argument2> <etc>
					// 0  1         2           3
					commandArguments := getCommands(rawLine)
					switch commandArguments[1] {
					case "import":
						{
							currentCommand = "import"
							options["file"] = commandArguments[2]
							break
						}
					default:
						{
							processedLines = append(processedLines, fmt.Sprintf("# Warning: Unknown command %s", commandArguments[1]))
						}
					}
				}
				processedLines = append(processedLines, rawLine)
			}
		}
	}

	joinedFilter := strings.Join(processedLines, "\n")
	return regexp.MustCompile(`\n{2,}`).ReplaceAllString(joinedFilter, "\n\n"), errorList
}

func importBaseFilter(filterName string) (string, error) {
	path := filepath.Join(BASE_FILTERS_PATH, filterName)

	filterData, err := os.ReadFile(path)

	if err != nil {
		path = filepath.Join(THIRD_PARTY_FILTERS_PATH, filterName)
		filterData, err = os.ReadFile(path)
	}

	return string(filterData), err
}

func getCommands(rawCommand string) []string {
	re := regexp.MustCompile(`\s+`)

	return re.Split(rawCommand, -1)
}
