package block

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Timestamp     time.Time
	Data          string
	PrevBlockHash string
	BlockHash     string
	Index         int
	Nonce         int
	Difficulty    int
}

func (b *Block) IsBlockHashValid() bool {
	prefix := b.BlockHash[:b.Difficulty]
	for _, char := range prefix {
		if char != '0' {
			return false
		}
	}
	return true
}

func (b *Block) CalculateBlockHash() string {
	h := sha256.New()
	h.Write([]byte(b.String()))
	h.Write([]byte(fmt.Sprint(b.Nonce)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *Block) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", b.Timestamp, b.Data, b.PrevBlockHash, b.BlockHash, b.Index,
		b.Difficulty)
}
