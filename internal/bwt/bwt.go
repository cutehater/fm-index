package bwt

import (
	"errors"
	"index/suffixarray"
	"reflect"
	"sort"
)

var CheckEndSymbol = true

var ErrEndSymbolExisted = errors.New("bwt: end-symbol existed in string")

var ErrEmptySequence = errors.New("bwt: empty sequence")

func Transform(sequence []byte, endSymbol byte) ([]byte, error) {
	if len(sequence) == 0 {
		return nil, ErrEmptySequence
	}
	if CheckEndSymbol {
		for _, character := range sequence {
			if character == endSymbol {
				return nil, ErrEndSymbolExisted
			}
		}
	}
	suffixArray := SuffixArray(sequence)
	bwt, err := FromSuffixArray(sequence, suffixArray, endSymbol)
	return bwt, err
}

func InverseTransform(transformedSequence []byte, endSymbol byte) []byte {
	seqLen := len(transformedSequence)
	matrix := make([][]byte, seqLen)
	for i := 0; i < seqLen; i++ {
		matrix[i] = make([]byte, seqLen)
	}

	for i := 0; i < seqLen; i++ {
		for j := 0; j < seqLen; j++ {
			matrix[j][seqLen-1-i] = transformedSequence[j]
		}
		sort.Sort(Matrix(matrix))
	}

	res := make([]byte, seqLen-1)
	for _, row := range matrix {
		if row[seqLen-1] == endSymbol {
			res = row[0 : seqLen-1]
			break
		}
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

var ErrInvalidSuffixArray = errors.New("bwt: invalid suffix array")

func FromSuffixArray(sequence []byte, suffixArray []int, endSymbol byte) ([]byte, error) {
	if len(sequence) == 0 {
		return nil, ErrEmptySequence
	}
	if len(sequence)+1 != len(suffixArray) || suffixArray[0] != len(sequence) {
		return nil, ErrInvalidSuffixArray
	}
	bwt := make([]byte, len(suffixArray))
	bwt[0] = sequence[len(sequence)-1]
	for i := 1; i < len(suffixArray); i++ {
		if suffixArray[i] == 0 {
			bwt[i] = endSymbol
		} else {
			bwt[i] = sequence[suffixArray[i]-1]
		}
	}
	return bwt, nil
}
