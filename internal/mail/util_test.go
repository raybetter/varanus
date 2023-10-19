package mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeReverseChunks(t *testing.T) {

	type TestCase struct {
		start, end, chunkSize int
		expectedChunks        []chunk
	}

	testCases := []TestCase{
		{
			start:     1,
			end:       10,
			chunkSize: 1,
			expectedChunks: []chunk{
				{10, 10},
				{9, 9},
				{8, 8},
				{7, 7},
				{6, 6},
				{5, 5},
				{4, 4},
				{3, 3},
				{2, 2},
				{1, 1},
			},
		},
		{
			start:     1,
			end:       10,
			chunkSize: 2,
			expectedChunks: []chunk{
				{9, 10},
				{7, 8},
				{5, 6},
				{3, 4},
				{1, 2},
			},
		},
		{
			start:     1,
			end:       10,
			chunkSize: 3,
			expectedChunks: []chunk{
				{8, 10},
				{5, 7},
				{2, 4},
				{1, 1},
			},
		},
		{
			start:     12,
			end:       12,
			chunkSize: 3,
			expectedChunks: []chunk{
				{12, 12},
			},
		},
		//empty chunks because start and end are reversed
		{
			start:          12,
			end:            1,
			chunkSize:      3,
			expectedChunks: []chunk{},
		},
		//empty chunks because chunk size less than 1
		{
			start:          1,
			end:            10,
			chunkSize:      0,
			expectedChunks: []chunk{},
		},
		{
			start:          1,
			end:            10,
			chunkSize:      -1,
			expectedChunks: []chunk{},
		},
	}

	for index, testCase := range testCases {
		actualChunks := makeReverseChunks(testCase.start, testCase.end, testCase.chunkSize)
		assert.Equal(t, testCase.expectedChunks, actualChunks, "for test %d", index)
	}

}
