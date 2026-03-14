package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type PoeNinjaData struct {
	Leagues        []PoeLeague
	CurrencyPrices map[string]PoeNinjaCurrencyPrices
}

type PoeNinjaCurrencyPrices struct {
}

type PoeNinjaIndexStateItem struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	DisplayName string `json:"displayName"`
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
	IsHardcore  bool
	IsEternal   bool
}

var eternalLeagues = []string{"standard", "hardcore"}
var divinationCardsFilterPath = filepath.Join("base-filters", "full-stack-divination-cards.filter")

const MAX_CACHE_AGE = 2 * 60 * 60 // 2 hours in seconds

func GetIndexState() (map[string][]PoeNinjaIndexStateItem, error) {
	resp, err := http.Get("https://poe.ninja/poe1/api/data/index-state")

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
		isEternalLeague := slices.Contains(eternalLeagues, v.Url)

		poeLeagues = append(poeLeagues, PoeLeague{
			Name:        v.Name,
			DisplayName: v.DisplayName,
			Url:         v.Url,
			IsHardcore:  v.Url == "hardcore" || strings.HasSuffix(v.Url, "hc"),
			IsEternal:   isEternalLeague,
		})
	}

	return poeLeagues, nil
}

// @todo(nick-ng): the poe.ninja endpoint this calls doesn't return stack size anymore. That is in a different endpoint. you have to update this before you can use it again
func MakeDivinationCardsFilterPoeNinja() error {
	leagues, err := GetTradeChallengeLeagues()
	if err != nil {
		slog.Error("error getting challenge leagues", "err", err)
		return err
	}

	var softcoreLeague PoeLeague

	for _, league := range leagues {
		if !league.IsHardcore && !league.IsEternal {
			softcoreLeague = league
			break
		}
	}

	if softcoreLeague.Name == "" {
		slog.Error("error getting challenge leagues", "err", err)
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
		url := fmt.Sprintf("https://poe.ninja/poe1/api/economy/exchange/current/overview?league=%s&type=DivinationCard", softcoreLeague.Name)

		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("couldn't fetch divination cards")
			return errors.New("couldn't fetch divination cards")
		}

		resBody, err := io.ReadAll(resp.Body)

		if err != nil {
			slog.Error("error: couldn't get body", "err", err)
			return errors.New("couldn't get body")
		}

		err = json.Unmarshal(resBody, &poeNinjaDivinationCardsResponse)

		if err != nil {
			slog.Error("error decoding divination cards response", "err", err, "url", url)
			return err
		}

		poeNinjaDivinationCardsResponse.Timestamp = now.Unix()

		cacheData, err := json.Marshal(poeNinjaDivinationCardsResponse)

		if err == nil {
			err = os.WriteFile(cachePath, []byte(cacheData), 0666)

			if err != nil {
				slog.Error("error encoding divination cards to JSON", "err", err)
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
					"\t#!BrightBackground!# 255",
					"\t#!CurrencyBorder!# 230",
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
