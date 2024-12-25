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
	BWTLast                   []byte
	BWTFirst                  []byte
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

	suffixArray := bwt.SuffixArray(sequence)
	fmIndex.SuffixArray = suffixArray

	fmIndex.BWTLast, err = bwt.FromSuffixArray(sequence, fmIndex.SuffixArray, fmIndex.EndSymbol)
	if err != nil {
		return nil, err
	}

	firstColumn := make([]byte, len(sequence)+1)
	firstColumn[0] = fmIndex.EndSymbol
	for i := 1; i <= len(sequence); i++ {
		firstColumn[i] = sequence[suffixArray[i]]
	}
	fmIndex.BWTFirst = firstColumn

	letterCounts := make([]int, 128)
	for _, letter := range fmIndex.BWTLast {
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

	fmIndex.LexicallySmallerCharCount = computeLexicallySmallerCharCount(fmIndex.BWTFirst)
	fmIndex.Occurrences = computeOccurrences(fmIndex.BWTLast, fmIndex.Alphabet)

	return fmIndex.BWTLast, nil
}

func (fmIndex *FMIndex) LastToFirst(index int) int {
	letter := fmIndex.BWTLast[index]
	return fmIndex.LexicallySmallerCharCount[letter] + int(fmIndex.Occurrences[letter][index])
}

func (fmIndex *FMIndex) nextLetterInAlphabet(currentLetter byte) byte {
	for i, letter := range fmIndex.Alphabet {
		if letter == currentLetter {
			if i < len(fmIndex.Alphabet)-1 {
				return fmIndex.Alphabet[i+1]
			}
			return letter
		}
	}
	return 0
}

func (fmIndex *FMIndex) Locate(pattern []byte, checkMatchOnly bool) ([]int, error) {
	if len(pattern) == 0 {
		return nil, nil
	}

	for _, letter := range pattern {
		if fmIndex.LetterCounts[letter] == 0 {
			return nil, nil
		}
	}

	n := len(fmIndex.BWTLast)
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
	buffer.WriteString(fmt.Sprintf("BWTLast: %s\n", string(fmIndex.BWTLast)))
	buffer.WriteString(fmt.Sprintf("Alphabet: %s\n", string(fmIndex.Alphabet)))
	buffer.WriteString("First Column:\n")
	buffer.WriteString(string(fmIndex.BWTFirst) + "\n")
	buffer.WriteString("Lexically Smaller Character Count:\n")
	for _, letter := range fmIndex.Alphabet {
		buffer.WriteString(fmt.Sprintf("  %c: %d\n", letter, fmIndex.LexicallySmallerCharCount[letter]))
	}
	return buffer.String()
}

func computeLexicallySmallerCharCount(firstColumn []byte) []int {
	res := make([]int, 128)
	count := 0
	for _, c := range firstColumn {
		if res[c] == 0 {
			res[c] = count
		}
		count++
	}
	return res
}

func computeOccurrences(bwt []byte, letters []byte) [][]int32 {
	occurences := make([][]int32, 128)
	for _, letter := range letters {
		first := make([]int32, 1, len(bwt))
		first[0] = 0
		occurences[letter] = first
	}
	t := make([]int32, 1, len(bwt))
	t[0] = 1
	occurences[bwt[0]] = t
	var letter byte
	var key, letterInt int
	var values []int32
	for _, letter = range bwt[1:] {
		letterInt = int(letter)
		for key, values = range occurences {
			if values == nil {
				continue
			}

			if key == letterInt {
				values = append(values, values[len(values)-1]+1)
			} else {
				values = append(values, values[len(values)-1])
			}
		}
	}
	return occurences
}
