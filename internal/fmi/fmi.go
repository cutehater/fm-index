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
	FirstColumn               []byte
	Alphabet                  []byte
	LetterCounts              []int
	LexicallySmallerCharCount []int
	Occurrences               []*[]int32
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

	fmIndex.BWT, err = bwt.FromSuffixArray(sequence, fmIndex.SuffixArray, fmIndex.EndSymbol)
	if err != nil {
		return nil, err
	}

	firstColumn := make([]byte, len(sequence)+1)
	firstColumn[0] = fmIndex.EndSymbol
	for i := 1; i <= len(sequence); i++ {
		firstColumn[i] = sequence[suffixArray[i]]
	}
	fmIndex.FirstColumn = firstColumn

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

	fmIndex.LexicallySmallerCharCount = computeLexicallySmallerCharCount(fmIndex.FirstColumn)
	fmIndex.Occurrences = computeOccurrences(fmIndex.BWT, fmIndex.Alphabet)

	return fmIndex.BWT, nil
}

func (fmIndex *FMIndex) LastToFirst(index int) int {
	letter := fmIndex.BWT[index]
	return fmIndex.LexicallySmallerCharCount[letter] + int((*fmIndex.Occurrences[letter])[index])
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

		// Consider only exact matches
		start := fmIndex.LexicallySmallerCharCount[lastLetter]
		if currentMatch.start > 0 {
			start += int((*fmIndex.Occurrences[lastLetter])[currentMatch.start-1])
		}
		end := fmIndex.LexicallySmallerCharCount[lastLetter] + int((*fmIndex.Occurrences[lastLetter])[currentMatch.end]-1)

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

func (fmIndex *FMIndex) Match(pattern []byte) (bool, error) {
	if len(pattern) == 0 {
		return false, nil
	}

	for _, letter := range pattern {
		if fmIndex.LetterCounts[letter] == 0 {
			return false, nil
		}
	}

	n := len(fmIndex.BWT)
	var matches Stack
	matches.Put(sMatch{query: pattern, start: 0, end: n - 1})

	for !matches.Empty() {
		currentMatch := matches.Pop()
		remainingPattern := currentMatch.query[:len(currentMatch.query)-1]
		lastLetter := currentMatch.query[len(currentMatch.query)-1]

		start := fmIndex.LexicallySmallerCharCount[lastLetter]
		if currentMatch.start > 0 {
			start += int((*fmIndex.Occurrences[lastLetter])[currentMatch.start-1])
		}
		end := fmIndex.LexicallySmallerCharCount[lastLetter] + int((*fmIndex.Occurrences[lastLetter])[currentMatch.end]-1)

		if start > end {
			continue
		}

		if len(remainingPattern) == 0 {
			return true, nil
		}

		matches.Put(sMatch{query: remainingPattern, start: start, end: end})
	}

	return false, nil
}

func (fmIndex *FMIndex) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("EndSymbol: %c\n", fmIndex.EndSymbol))
	buffer.WriteString(fmt.Sprintf("BWT: %s\n", string(fmIndex.BWT)))
	buffer.WriteString(fmt.Sprintf("Alphabet: %s\n", string(fmIndex.Alphabet)))
	buffer.WriteString("First Column:\n")
	buffer.WriteString(string(fmIndex.FirstColumn) + "\n")
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

func computeOccurrences(bwt []byte, letters []byte) []*[]int32 {
	if letters == nil {
		count := make([]int, 128)
		for _, b := range bwt {
			if count[b] == 0 {
				count[b]++
			}
		}

		letters = make([]byte, 0, 128)
		for b, c := range count {
			if c > 0 {
				letters = append(letters, byte(b))
			}
		}
	}

	occurences := make([]*[]int32, 128)
	for _, letter := range letters {
		t := make([]int32, 1, len(bwt))
		t[0] = 0
		occurences[letter] = &t
	}
	t := make([]int32, 1, len(bwt))
	t[0] = 1
	occurences[bwt[0]] = &t
	var letter byte
	var k, letterInt int
	var v *[]int32
	for _, letter = range bwt[1:] {
		letterInt = int(letter)
		for k, v = range occurences {
			if v == nil {
				continue
			}

			if k == letterInt {
				*v = append(*v, (*v)[len(*v)-1]+1)
			} else {
				*v = append(*v, (*v)[len(*v)-1])
			}
		}
	}
	return occurences
}
