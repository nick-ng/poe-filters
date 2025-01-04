package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"poe-filters/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ProcessedFilterFlags struct {
	Game string // poe1, poe2, or unset
}

const MY_FILTERS_PATH string = "my-filters"
const BASE_FILTERS_PATH string = "base-filters"
const BUILD_FILTERS_PATH string = "build-filters"
const THIRD_PARTY_FILTERS_PATH string = "third-party-filters"
const OUTPUT_FILTERS_PATH string = "output-filters"
const CACHE_PATH string = "cache"

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

	utils.MakeDivinationCardsFilterPoeNinja()

	utils.MkDirIfNotExist(MY_FILTERS_PATH)
	utils.MkDirIfNotExist(OUTPUT_FILTERS_PATH)
	utils.MkDirIfNotExist(BASE_FILTERS_PATH)
	utils.MkDirIfNotExist(THIRD_PARTY_FILTERS_PATH)
	utils.MkDirIfNotExist(CACHE_PATH)

	path1 := filepath.Join(MY_FILTERS_PATH)

	dat1, err := os.ReadDir(path1)

	if err != nil {
		fmt.Println("Error reading file", err)
	}

	// parse poe2 filters first
	var poe1Filters []string
	var poe2Filters []string
	for _, dir := range dat1 {
		filterName := dir.Name()
		if strings.Contains(filterName, "poe2") {
			poe2Filters = append(poe2Filters, filterName)
		} else {
			poe1Filters = append(poe1Filters, filterName)
		}
	}

	allFilters := append(poe2Filters, poe1Filters...)

	fmt.Printf("Found %d filters\n", len(allFilters))

	for i, filterName := range allFilters {
		// if err != nil {
		// 	continue
		// }

		// filterName := dir.Name()

		fmt.Printf("%d: %s... ", i+1, filterName)

		path := filepath.Join(MY_FILTERS_PATH, filterName)
		filter, flags, errList := processFilter(path, false)

		outputFilterPath := filepath.Join(OUTPUT_FILTERS_PATH, filterName)
		gameFilterPath := filepath.Join(homeDir, "Documents", "My Games", "Path of Exile", filterName)

		if flags.Game == "poe2" {
			gameFilterPath = filepath.Join(homeDir, "Documents", "My Games", "Path of Exile 2", filterName)
		}

		if len(filter) == 0 {
			err := os.Remove(outputFilterPath)
			err2 := os.Remove(gameFilterPath)

			if err != nil && !os.IsNotExist(err) {
				fmt.Print(err)
			}

			if err2 != nil && !os.IsNotExist(err2) {
				fmt.Print(err2)
			}

			fmt.Println("skipped")

			continue
		}

		filterData := []byte(filter)

		err := os.WriteFile(outputFilterPath, filterData, 0666)

		if err != nil {
			fmt.Println("\nError writing filter to output-filters", filterName, err)
			continue
		}

		if filterName != "example.filter" && filterName != "example2.filter" {
			err = os.WriteFile(gameFilterPath, filterData, 0666)

			if err != nil {
				fmt.Println("\nError writing filter to PoE Directory", filterName, err)
				continue
			}
		}

		if len(errList) > 1 {
			fmt.Printf("done with %d errors\n", len(errList))
			continue
		} else if len(errList) == 1 {
			fmt.Println("done with 1 error")
			continue
		} else {
			fmt.Println("done")
		}
	}

	utils.PrintSoundStats()
}

// @todo(nick-ng): move some functions to separate files
func processFilter(filterPath string, isImported bool) (string, ProcessedFilterFlags, []error) {
	var errorList []error

	flags := ProcessedFilterFlags{}

	filterData, err := os.ReadFile(filterPath)

	if err != nil {
		return "", flags, append(errorList, err)
	}

	rawLines := strings.Split(string(filterData), "\n")

	var filterChunks []string
	if !isImported {
		filterChunks = append(filterChunks, `Show
	#!DefaultBackground!# 120
	Continue`)
	}

	var currentCommand string
	var importAfterTimestamp int64
	now := time.Now().Unix()
	options := make(map[string][]string)
	// handle import command (multi-line)
	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)
		switch currentCommand {
		case "import":
			{
				subCommandArguments := getCommands(trimmedLine)
				switch subCommandArguments[0] {
				case "exclude":
					fallthrough
				case "del":
					fallthrough
				case "delete":
					{
						regexpString := strings.Join(subCommandArguments[1:], " ")
						options["delete"] = append(options["delete"], regexpString)
					}
				case "include":
					{
						regexpString := strings.Join(subCommandArguments[1:], " ")
						options["include"] = append(options["delete"], regexpString)
					}
				case "maxarea":
					{
						options["maxarea"] = []string{subCommandArguments[1]}
					}
				case "import":
					fallthrough
				case "noop": // the line didn't start with #!
					{
						// why is this "if" statement necessary? doesn't it match both cases?
						if !strings.HasPrefix(trimmedLine, "#") || subCommandArguments[0] == "import" {
							// it's a non-comment line so import the filter here then write the line
							_, present := options["file"]
							if present {
								_, fullPath, err := importBaseFilter(options["file"][0])

								if err != nil {
									errorList = append(errorList, err)
									filterChunks = append(filterChunks, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									filterChunks = append(filterChunks, fmt.Sprintf("#?  %s\n", err))
								}

								tempFilter, _, errs := processFilter(fullPath, true)

								if len(errs) != 0 {
									errorList = append(errorList, err)
									filterChunks = append(filterChunks, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									filterChunks = append(filterChunks, fmt.Sprintf("#?  %s\n", err))
								} else if now < int64(importAfterTimestamp) {
									filterChunks = append(filterChunks, fmt.Sprintf("#?  skipped because generated at %d", now))
								} else { // nothing wrong. we can modify the imported filter
									for _, deleteRegexpString := range options["delete"] {
										deleteRegexp, err := regexp.Compile(deleteRegexpString)

										if err != nil {
											errorString := fmt.Sprintf("couldn't compile regexp: %s", deleteRegexpString)
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter = deleteRegexp.ReplaceAllString(tempFilter, "")
										}
									}

									for _, includeRegexpString := range options["include"] {
										includeRegexp, err := regexp.Compile(includeRegexpString)

										if err != nil {
											errorString := fmt.Sprintf("couldn't compile regexp: %s", includeRegexpString)
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter = includeRegexp.ReplaceAllString(tempFilter, "")
										}
									}

									if options["maxarea"] != nil && len(options["maxarea"]) > 0 {
										maxarea, err := strconv.ParseInt(options["maxarea"][0], 10, 0)

										if err != nil {
											errorString := fmt.Sprintf("maxarea argument not an integer: %s", options["maxarea"][0])
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter2, err := utils.LimitMaxAreaLevel(tempFilter, int(maxarea))

											if err != nil {
												errorString := fmt.Sprintf("couldn't lower filter's level: %s", err)
												errorList = append(errorList, errors.New(errorString))
												filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
											} else {
												tempFilter = tempFilter2
											}
										}
									}

									filterChunks = append(filterChunks, tempFilter)
									filterChunks = append(filterChunks, fmt.Sprintf("#? end of %s\n", options["file"]))
								}

							} else {
								filterChunks = append(filterChunks, "#? error: No file specified.\n")
							}

							currentCommand = ""
							clear(options)
						}

						if subCommandArguments[0] == "import" {
							currentCommand = "import"
							options["file"] = []string{subCommandArguments[1]}

							tempLine := utils.CleanUpCommand(rawLine)
							filterChunks = append(filterChunks, tempLine)
						} else {
							filterChunks = append(filterChunks, rawLine)
						}
					}
				default:
					{
						filterChunks = append(filterChunks, trimmedLine)

						warning := fmt.Sprintf("#? warning: Unknown sub-command %s", subCommandArguments[0])
						filterChunks = append(filterChunks, warning)
					}
				}
			}
		// there is no current command or the current line isn't a sub command of the current command
		default:
			{
				commandArguments := getCommands(trimmedLine)
				switch commandArguments[0] {
				case "poe2":
					{
						// when a filter is imported, the game flag is ignored
						flags.Game = "poe2"
						tempLine := utils.CleanUpCommand(rawLine)
						filterChunks = append(filterChunks, tempLine)
					}
				case "import":
					{
						currentCommand = "import"
						options["file"] = []string{commandArguments[1]}
						if len(commandArguments) > 2 {
							importAfterTimestamp, err = strconv.ParseInt(commandArguments[2], 10, 0)

							if err != nil {
								errorString := fmt.Sprintf("couldn't parse import after timestamp: %s", commandArguments[2])
								errorList = append(errorList, errors.New(errorString))
							}
						}

						tempLine := utils.CleanUpCommand(rawLine)
						filterChunks = append(filterChunks, tempLine)
					}
				case "del":
					fallthrough
				case "delete":
					{
						filterChunks = append(filterChunks, rawLine)
						warning := fmt.Sprintf("#? warning: %s only allowed during an import", commandArguments[0])
						filterChunks = append(filterChunks, warning)
					}
				case "skip":
					{
						return "", flags, []error{}
					}
				case "noop":
					fallthrough
				default:
					{
						filterChunks = append(filterChunks, rawLine)
						// Handle in the next loop so we don't get double warnings
						// noop is also here
					}
				}
			}
		}
	}

	tempFilter := strings.Join(filterChunks, "\n")
	tempFilterLines := strings.Split(tempFilter, "\n")
	// handle import command (multi-line)
	var tempFilterChunks []string
	var rawCommand string
	var isCommandStarted bool
	clear(options)
	for _, line := range tempFilterLines {
		trimmedLine := strings.TrimSpace(line)
		commandArgs := getCommands(trimmedLine)
		command := strings.ToLower(commandArgs[0])
		if isCommandStarted && command != "custom" && command != "custombig" {
			isCommandStarted = false
			dropLevelChunk, err := utils.GetDropLevelFilter(rawCommand, options["custom"], options["custombig"])
			if err != nil {
				tempFilterChunks = append(tempFilterChunks, fmt.Sprintf("#? error: couldn't make droplevel filter from \"%s\" because %s", rawCommand, err))
			} else {
				tempFilterChunks = append(tempFilterChunks, dropLevelChunk)
			}
		}

		completedCommand := utils.CleanUpCommand(line)
		switch command {
		case "droplevel":
			{
				isCommandStarted = true
				rawCommand = line
				clear(options)
				tempFilterChunks = append(tempFilterChunks, completedCommand)
			}
		case "custom":
			fallthrough
		case "custombig":
			{
				customStyle := strings.Join(commandArgs[1:], " ")
				options[command] = append(options[command], customStyle)
				tempFilterChunks = append(tempFilterChunks, completedCommand)
			}
		default:
			{
				tempFilterChunks = append(tempFilterChunks, line)
			}
		}
	}

	filterChunks = tempFilterChunks

	// @todo(nick-ng): move these to a method so we can process all commands (multi-line or otherwise) in a single loop?
	// separate loop for all non-multi-line commands. if we encounter a
	// different command during a multi-line command, we need to handle the
	// multi-line command and the new command.
	tempFilter = strings.Join(filterChunks, "\n")
	tempFilterLines = strings.Split(tempFilter, "\n")
	clear(tempFilterChunks)
	for _, rawLine := range tempFilterLines {
		trimmedLine := strings.TrimSpace(rawLine)

		if strings.HasPrefix(trimmedLine, "#! ") {
			commandArguments := getCommands(trimmedLine)
			switch commandArguments[0] {
			case "import":
				fallthrough
			case "del":
				fallthrough
			case "delete":
				fallthrough
			case "noop":
				{
					tempFilterChunks = append(tempFilterChunks, rawLine)
					// noop
				}
			case "links":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					filterBlock, err := utils.GetSocketGroupFilter(commandArguments[1], commandArguments[2:]...)

					if err != nil {
						e := fmt.Sprintf("#? error: couldn't generate links\n#?  %s", err)
						tempFilterChunks = append(tempFilterChunks, e)
					} else {
						tempFilterChunks = append(tempFilterChunks, filterBlock)
					}

				}
			case "linksa":
				fallthrough
			case "linksarmor":
				fallthrough
			case "linksarmour":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					newLine := utils.GetArmourSocketGroupFilter(commandArguments[1], commandArguments[2:]...)
					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			case "weapon":
				fallthrough
			case "weapons":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)
					maxLevel := 99
					minLevel := 0
					seriousAdjustment := 1
					showAdjustment := 7

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least the weapon group (slow-wands, crit-bows, etc.) to work")
						continue
					}

					if len(commandArguments) >= 3 {
						maxLevel64, err := strconv.ParseInt(commandArguments[2], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse max level")
						} else {
							maxLevel = int(maxLevel64)
						}
					}

					if len(commandArguments) >= 4 {
						minLevel64, err := strconv.ParseInt(commandArguments[3], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse min level")
						} else {
							minLevel = int(minLevel64)
						}
					}

					if len(commandArguments) >= 5 {
						seriousAdjustment64, err := strconv.ParseInt(commandArguments[4], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse serious adjustment")
						} else {
							seriousAdjustment = int(seriousAdjustment64)
						}
					}

					if len(commandArguments) >= 6 {
						showAdjustment64, err := strconv.ParseInt(commandArguments[5], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse show adjustment")
						} else {
							showAdjustment = int(showAdjustment64)
						}
					}

					newLine, err := utils.GetWeaponGroupFilter(commandArguments[1], maxLevel, minLevel, seriousAdjustment, showAdjustment)

					if err != nil {
						tempFilterChunks = append(tempFilterChunks, "#? warning: couldn't get weapon group")
						continue
					}

					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			// // @todo(nick-ng): move this to its own loop
			// case "droplevel":
			// 	{
			// 		// makes item filter based on item drop level
			// 		_ = utils.ParseFlags(rawLine)
			// 		newLine, err := utils.GetDropLevelFilter(rawLine)

			// 		if err != nil {
			// 			tempFilterChunks = append(tempFilterChunks, "#? warning: couldn't get drop level filter")
			// 			continue
			// 		}

			// 		tempFilterChunks = append(tempFilterChunks, newLine)
			// 	}
			case "tts":
				{
					newLine := utils.MakeTts(trimmedLine)
					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			default:
				{
					warning := fmt.Sprintf("#? warning: Unknown command %s", commandArguments[0])
					tempFilterChunks = append(tempFilterChunks, warning)
				}
			}
		} else {
			tempFilterChunks = append(tempFilterChunks, rawLine)
		}
	}

	filterChunks = tempFilterChunks
	filter := strings.Join(filterChunks, "\n")

	if !isImported {
		filter = utils.ApplyAllTokens(filter)
		filter = utils.CleanUpFilter(filter)
	}

	if strings.HasSuffix(filterPath, ".ruthlessfilter") {
		filter = strings.ReplaceAll(filter, "Hide", "Show")
	}

	return filter, flags, errorList
}

func importBaseFilter(filterName string) (string, string, error) {
	var err error

	for _, directoryName := range []string{BASE_FILTERS_PATH, BUILD_FILTERS_PATH, THIRD_PARTY_FILTERS_PATH} {
		path := filepath.Join(directoryName, filterName)

		filterData, err := os.ReadFile(path)

		if err == nil {
			return string(filterData), path, err
		}
	}

	return "", "", err
}

func getCommands(rawCommand string) []string {
	commands := strings.Split(rawCommand, " ")
	if commands[0] != "#!" {
		return []string{"noop"}
	}

	return commands[1:]
}
