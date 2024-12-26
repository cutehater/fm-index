package fmi

import (
	"bytes"
	"fm-index/internal/bwt"
	"fmt"
	"sort"
)

type FMIndex struct {
	EndSymbol                 byte
	SuffixArray               []int
	BWT                       []byte
	Alphabet                  []byte
	LetterCounts              []int
	LexicallySmallerCharCount []int
	Occurrences               [][]int32
}

func NewFMIndex() *FMIndex {
	fmIndex := new(FMIndex)
	fmIndex.EndSymbol = byte(0)
	return fmIndex
}

func (fmIndex *FMIndex) Transform(sequence []byte) ([]byte, error) {
	if len(sequence) == 0 {
		return nil, bwt.ErrEmptySequence
	}
	var err error

	fmIndex.SuffixArray = bwt.GetSuffixArray(sequence)

	fmIndex.BWT, err = bwt.Transform(sequence, fmIndex.EndSymbol)
	if err != nil {
		return nil, err
	}

	letterCounts := make([]int, 128)
	for _, letter := range fmIndex.BWT {
		letterCounts[letter]++
	}
	letterCounts[fmIndex.EndSymbol] = 0
	fmIndex.LetterCounts = letterCounts

	alphabet := make([]byte, 0, 128)
	for letter, count := range letterCounts {
		if count > 0 {
			alphabet = append(alphabet, byte(letter))
		}
	}
	fmIndex.Alphabet = alphabet

	fmIndex.LexicallySmallerCharCount = computeLexicallySmallerCharCount(fmIndex.BWT)
	fmIndex.Occurrences = computeOccurrences(fmIndex.BWT, fmIndex.Alphabet)

	return fmIndex.BWT, nil
}

func (fmIndex *FMIndex) Locate(pattern []byte) ([]int, error) {
	if len(pattern) == 0 {
		return nil, nil
	}

	for _, letter := range pattern {
		if fmIndex.LetterCounts[letter] == 0 {
			return nil, nil
		}
	}

	n := len(fmIndex.BWT)
	var matches Stack
	matches.Put(sMatch{query: pattern, start: 0, end: n - 1})

	locationsMap := make(map[int]struct{})

	for !matches.Empty() {
		currentMatch := matches.Pop()
		remainingPattern := currentMatch.query[:len(currentMatch.query)-1]
		lastLetter := currentMatch.query[len(currentMatch.query)-1]

		start := fmIndex.LexicallySmallerCharCount[lastLetter]
		if currentMatch.start > 0 {
			start += int(fmIndex.Occurrences[lastLetter][currentMatch.start-1])
		}
		end := fmIndex.LexicallySmallerCharCount[lastLetter] + int(fmIndex.Occurrences[lastLetter][currentMatch.end]-1)

		if start > end {
			continue
		}

		if len(remainingPattern) == 0 {
			for _, location := range fmIndex.SuffixArray[start : end+1] {
				locationsMap[location] = struct{}{}
			}
		} else {
			matches.Put(sMatch{query: remainingPattern, start: start, end: end})
		}
	}

	locations := make([]int, 0, len(locationsMap))
	for location := range locationsMap {
		locations = append(locations, location)
	}
	sort.Ints(locations)
	return locations, nil
}

func (fmIndex *FMIndex) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("EndSymbol: %c\n", fmIndex.EndSymbol))
	buffer.WriteString(fmt.Sprintf("BWT: %s\n", string(fmIndex.BWT)))
	buffer.WriteString(fmt.Sprintf("Alphabet: %s\n", string(fmIndex.Alphabet)))
	buffer.WriteString("First Column:\n")
	buffer.WriteString("Lexically Smaller Character Count:\n")
	for _, letter := range fmIndex.Alphabet {
		buffer.WriteString(fmt.Sprintf("  %c: %d\n", letter, fmIndex.LexicallySmallerCharCount[letter]))
	}
	return buffer.String()
}

func computeLexicallySmallerCharCount(bwt []byte) []int {
	res := make([]int, 128)
	for _, c := range bwt {
		res[c]++
	}
	for i := 1; i < len(res); i++ {
		res[i] = res[i] + res[i-1]
	}
	for i := len(res) - 1; i > 0; i-- {
		res[i] = res[i-1]
	}
	res[0] = 0
	return res
}

func computeOccurrences(bwt []byte, letters []byte) [][]int32 {
	occurences := make([][]int32, 128)
	for i := range occurences {
		occurences[i] = make([]int32, len(bwt))
	}

	for i, letter := range bwt {
		if i > 0 {
			for j := range occurences {
				occurences[j][i] = occurences[j][i-1]
			}
		}
		occurences[letter][i]++
	}

	return occurences
}
