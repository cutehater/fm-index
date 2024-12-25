package bwt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformAndInverseTransform(t *testing.T) {
	s := "abracadabra"
	trans := "ard$rcaaaabb"
	tr, err := Transform([]byte(s), '$')

	require.NoError(t, err)
	require.Equal(t, trans, string(tr))
	require.Equal(t, s, string(InverseTransform([]byte(trans), '$')))
}

func TestFromSuffixArray(t *testing.T) {
	s := "banana"
	trans := "annb$aa"

	sa := SuffixArray([]byte(s))
	B, err := FromSuffixArray([]byte(s), sa, '$')
	require.NoError(t, err)
	require.Equal(t, trans, string(B))
}

func TestFromSuffixArrayEmptySeq(t *testing.T) {
	s := ""

	sa := SuffixArray([]byte(s))
	_, err := FromSuffixArray([]byte(s), sa, '$')
	require.Error(t, err)
}

func TestSA(t *testing.T) {
	s := "raskolbasras"
	sa := SuffixArray([]byte(s))
	sa1 := []int{12, 10, 1, 7, 6, 3, 5, 4, 9, 0, 11, 2, 8}
	require.Equal(t, sa1, sa)
}
