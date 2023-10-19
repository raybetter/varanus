package mail

import (
	"strings"

	"github.com/emersion/go-imap"
)

type chunk struct {
	start int
	end   int
}

func makeReverseChunks(sequenceStart int, sequenceEnd int, chunkSize int) []chunk {
	if sequenceEnd < sequenceStart {
		return []chunk{}
	}
	if chunkSize < 1 {
		return []chunk{}
	}

	chunks := make([]chunk, 0, int((sequenceEnd-sequenceStart)/chunkSize)+1)

	current := sequenceEnd
	for current >= sequenceStart {
		end := current
		start := current - chunkSize + 1
		if start < sequenceStart {
			start = sequenceStart
		}
		chunks = append(chunks, chunk{start, end})
		current = start - 1
	}

	return chunks
}

func addressesToString(addresses []*imap.Address) string {
	if len(addresses) == 0 {
		return "unknown address"
	}
	addressStrings := make([]string, 0, len(addresses))
	for _, address := range addresses {
		addressStrings = append(addressStrings, address.Address())
	}
	return strings.Join(addressStrings, ", ")
}
