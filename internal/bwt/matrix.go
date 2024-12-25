package bwt

import "bytes"

type Matrix [][]byte

func (s Matrix) Len() int { return len(s) }
func (s Matrix) Less(i, j int) bool {
	return bytes.Compare(s[i], s[j]) < 0
}
func (s Matrix) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
