package fmi

import (
	"fm-index/internal/bwt"
	"github.com/stretchr/testify/require"
	"testing"
)

type Case struct {
	text    string
	pattern string
	want    []int
}

var cases = []Case{
	{"", "abc", []int{}},
	{"mississippi", "", []int{}},
	{"mississippi", "iss", []int{1, 4}},
	{"abcabcabc", "abc", []int{0, 3, 6}},
	{"abcabcabc", "gef", []int{}},
	{"abcabcabc", "gef", []int{}},
	{"abcabcabc", "xef", []int{}},
	{"acctatac", "ac", []int{0, 6}},
	{"acctatac", "tac", []int{5}},
	{"acctatac", "atac", []int{4}},
	{"acctatac", "acctatac", []int{0}},
}

func TestLocate(t *testing.T) {
	var err error
	var locations []int
	var fmi *FMIndex

	for _, c := range cases {
		fmi = NewFMIndex()
		_, err = fmi.Transform([]byte(c.text))

		if c.text != "" {
			require.NoError(t, err)
		} else {
			require.Equal(t, bwt.ErrEmptySequence, err)
			continue
		}

		locations, err = fmi.Locate([]byte(c.pattern))
		require.NoError(t, err)
		if len(locations) > 0 || len(c.want) > 0 {
			require.Equal(t, c.want, locations)
		}
	}
}
