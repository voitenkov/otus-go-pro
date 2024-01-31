package hw03frequencyanalysis

import (
	"sort"
	"strings"
	"unicode"
)

type dictionary struct {
	word  string
	count int
}

func Top10(input string) []string {
	// Place your code here.

	wordsSlice := make([]string, 0)

	if len(input) == 0 {
		return wordsSlice
	}

	words := strings.Fields(input)
	dictionaryMap := make(map[string]int)
	dictionarySlice := make([]dictionary, 0)
	var wordCount dictionary
	var wordPrepared string
	topListLen := 10

	for _, word := range words {
		wordPrepared = strings.ToLower(strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		}))

		if len(wordPrepared) > 0 {
			dictionaryMap[wordPrepared]++
		}
	}

	for key, value := range dictionaryMap {
		dictionarySlice = append(dictionarySlice, dictionary{key, value})
	}

	sort.Slice(dictionarySlice, func(i, j int) bool {
		if dictionarySlice[i].count != dictionarySlice[j].count {
			return dictionarySlice[i].count > dictionarySlice[j].count
		}
		return dictionarySlice[i].word < dictionarySlice[j].word
	})

	if len(dictionarySlice) < 10 {
		topListLen = len(dictionarySlice)
	}

	for i := 0; i < topListLen; i++ {
		wordCount = dictionarySlice[i]
		// fmt.Printf("%s, %d\n", wordCount.word, wordCount.count)
		wordsSlice = append(wordsSlice, wordCount.word)
	}

	return wordsSlice
}
