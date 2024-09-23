package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type PoeNinjaIndexStateItem struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	DisplayName string `json:"displayName"`
	Hardcore    bool   `json:"hardcore"`
	Indexed     bool   `json:"indexed"`
}

type PoeNinjaDivinationCardItem struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Icon         string  `json:"icon"`
	BaseType     string  `json:"baseType"`
	StackSize    int     `json:"stackSize"`
	ChaosValue   float64 `json:"chaosValue"`
	ExaltedValue float64 `json:"exaltedValue"`
	DivineValue  float64 `json:"divineValue"`
}

type PoeNinjaDivinationCardsResponse struct {
	Lines     []PoeNinjaDivinationCardItem `json:"lines"`
	Timestamp int64                        `json:"timestamp"`
}

type PoeLeague struct {
	Name        string
	DisplayName string
	Url         string
	Hardcore    bool
}

var permanentLeagues = []string{"standard", "hardcore"}
var divinationCardsFilterPath = filepath.Join("base-filters", "full-stack-divination-cards.filter")

const MAX_CACHE_AGE = 2 * 60 * 60 // 2 hours in seconds

func GetIndexState() (map[string][]PoeNinjaIndexStateItem, error) {
	resp, err := http.Get("https://poe.ninja/api/data/getindexstate")

	if err != nil {
		fmt.Println("couldn't fetch poe.ninja stats", err)
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("couldn't get body", err)
		return nil, err
	}

	poeNinjaStateIndex := map[string][]PoeNinjaIndexStateItem{}

	json.Unmarshal(resBody, &poeNinjaStateIndex)

	return poeNinjaStateIndex, nil
}

func GetTradeChallengeLeagues() ([]PoeLeague, error) {
	poeLeagues := []PoeLeague{}

	poeNinjaStateIndex, err := GetIndexState()

	if err != nil {
		fmt.Println("couldn't get league names")
		return nil, err
	}

	for _, v := range poeNinjaStateIndex["economyLeagues"] {
		isPermanentLeague := true

		for _, permanentLeague := range permanentLeagues {
			if permanentLeague == strings.ToLower(v.DisplayName) {
				isPermanentLeague = false
				break
			}
		}

		if isPermanentLeague {
			poeLeagues = append(poeLeagues, PoeLeague{
				Name:        v.Name,
				DisplayName: v.DisplayName,
				Url:         v.Url,
				Hardcore:    v.Hardcore,
			})
		}
	}

	return poeLeagues, nil
}

func MakeDivinationCardsFilterPoeNinja() error {
	leagues, err := GetTradeChallengeLeagues()

	if err != nil {
		return err
	}

	var softcoreLeague PoeLeague

	for _, league := range leagues {
		if !league.Hardcore {
			softcoreLeague = league
		}
	}

	if softcoreLeague.Name == "" {
		return errors.New("no softcore trade league found")
	}

	cachePath := filepath.Join("cache", strings.ToLower(fmt.Sprintf("poeninja-divination-cards-%s.json", softcoreLeague.Name)))

	cacheData, err := os.ReadFile(cachePath)

	poeNinjaDivinationCardsResponse := PoeNinjaDivinationCardsResponse{}

	now := time.Now()

	needRefetch := true

	if err == nil {
		json.Unmarshal(cacheData, &poeNinjaDivinationCardsResponse)

		if (poeNinjaDivinationCardsResponse.Timestamp - now.Unix()) < MAX_CACHE_AGE {
			needRefetch = false
		}
	}

	if needRefetch {
		url := fmt.Sprintf("https://poe.ninja/api/data/itemoverview?league=%s&type=DivinationCard", softcoreLeague.Name)

		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("couldn't fetch divination cards")
			return errors.New("couldn't fetch divination cards")
		}

		resBody, err := io.ReadAll(resp.Body)

		if err != nil {
			fmt.Println("couldn't get body")
			return errors.New("couldn't get body")
		}

		err = json.Unmarshal(resBody, &poeNinjaDivinationCardsResponse)

		if err != nil {
			return err
		}

		poeNinjaDivinationCardsResponse.Timestamp = now.Unix()

		cacheData, err := json.Marshal(poeNinjaDivinationCardsResponse)

		if err == nil {
			err = os.WriteFile(cachePath, []byte(cacheData), 0666)

			if err != nil {
				return err
			}
		}
	}

	divinationCardsByStackSize := make(map[int][]string)

	maxStackSize := 0

	for _, divinationCard := range poeNinjaDivinationCardsResponse.Lines {
		stackSize := divinationCard.StackSize

		divinationCardsByStackSize[stackSize] = append(divinationCardsByStackSize[stackSize], divinationCard.BaseType)

		if stackSize > maxStackSize {
			maxStackSize = stackSize
		}
	}

	var filterLines []string

	for i := 0; i <= maxStackSize; i++ {
		divinationCardNames := divinationCardsByStackSize[i]

		slices.Sort(divinationCardNames)

		if len(divinationCardNames) > 0 {
			filterLines =
				append(filterLines,
					"Show",
					fmt.Sprintf("\tBaseType == \"%s\"", strings.Join(divinationCardNames, "\" \"")),
					"\tClass \"Divination\"",
					"\tSetFontSize 45",
					"\t#!BrightBackground!#",
					"\t#!CurrencyBorder!#",
					"\tMinimapIcon 0 Pink UpsideDownHouse",
					"\tCustomAlertSound \"sounds/intel-thanks-steve.mp3\"")

			if i > 1 {
				filterLines = append(filterLines, fmt.Sprintf("\tStackSize >= %d", i))
			}

			filterLines = append(filterLines, "")
		}
	}

	filter := strings.Join(filterLines, "\n")

	os.WriteFile(divinationCardsFilterPath, []byte(filter), 0666)

	fmt.Printf("%d divination cards found\n", len(poeNinjaDivinationCardsResponse.Lines))

	return nil
}
