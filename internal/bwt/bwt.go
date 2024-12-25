package bwt

import (
	"errors"
	"index/suffixarray"
	"reflect"
	"sort"
)

var (
	ErrEmptySequence      = errors.New("bwt: empty sequence")
	ErrEndSymbolExisted   = errors.New("bwt: end-symbol existed in string")
	ErrInvalidSuffixArray = errors.New("bwt: invalid suffix array")
)

func Transform(sequence []byte, endSymbol byte) ([]byte, error) {
	if len(sequence) == 0 {
		return nil, ErrEmptySequence
	}
	for _, character := range sequence {
		if character == endSymbol {
			return nil, ErrEndSymbolExisted
		}
	}

	suffixArray := SuffixArray(sequence)
	bwt, err := FromSuffixArray(sequence, suffixArray, endSymbol)
	return bwt, err
}

func InverseTransform(transformedSequence []byte, endSymbol byte) []byte {
	seqLen := len(transformedSequence)

	counts := make(map[byte]int)
	for _, ch := range transformedSequence {
		counts[ch]++
	}

	chars := make([]byte, 0, len(counts))
	for ch := range counts {
		chars = append(chars, ch)
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })

	startPositions := make(map[byte]int)
	pos := 0
	for _, ch := range chars {
		startPositions[ch] = pos
		pos += counts[ch]
	}

	next := make([]int, seqLen)
	occurrences := make(map[byte]int)
	for i, ch := range transformedSequence {
		next[i] = startPositions[ch] + occurrences[ch]
		occurrences[ch]++
	}

	res := make([]byte, seqLen-1)
	index := 0
	for i := seqLen - 2; i >= 0; i-- {
		res[i] = transformedSequence[index]
		index = next[index]
	}

	return res
}

func SuffixArray(sequence []byte) []int {
	suffixArray := suffixarray.New(sequence)
	tmp := reflect.ValueOf(suffixArray).Elem().FieldByName("sa").FieldByIndex([]int{0})
	res := make([]int, len(sequence)+1)
	res[0] = len(sequence)
	for i := 0; i < len(sequence); i++ {
		res[i+1] = int(tmp.Index(i).Int())
	}
	return res
}

func FromSuffixArray(sequence []byte, suffixArray []int, endSymbol byte) ([]byte, error) {
	if len(sequence) == 0 {
		return nil, ErrEmptySequence
	}
	if len(sequence)+1 != len(suffixArray) || suffixArray[0] != len(sequence) {
		return nil, ErrInvalidSuffixArray
	}

	bwt := make([]byte, len(suffixArray))
	for i := 0; i < len(suffixArray); i++ {
		if suffixArray[i] == 0 {
			bwt[i] = endSymbol
		} else {
			bwt[i] = sequence[suffixArray[i]-1]
		}
	}
	return bwt, nil
}
