package block

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlock_CalculateBlockHash(t *testing.T) {
	block := &Block{
		Timestamp:     time.Time{},
		Data:          "Test",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         42,
		Difficulty:    0,
	}

	hash := block.CalculateBlockHash()
	assert.Equal(t, len(hash), 64, "Hash must be valid SHA256")
}

func TestBlock_IsBlockHashValid(t *testing.T) {

	block := &Block{
		Timestamp:     time.Time{},
		Data:          "Test",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         42,
		Difficulty:    4,
	}
	block.BlockHash = block.CalculateBlockHash()

	// Change the block hash to be valid
	goodBlockHash := []byte(block.BlockHash)
	goodBlockHash = append([]byte{'0', '0', '0', '0'}, goodBlockHash[4:]...)
	block.BlockHash = string(goodBlockHash)
	assert.True(t, block.IsBlockHashValid(), "Block hash should be prefixed with 0s equal to the difficulty")

	// Change the block hash to be invalid
	badBlockHash := []byte(block.BlockHash)
	badBlockHash[3] = byte('A')
	block.BlockHash = string(badBlockHash)
	assert.False(t, block.IsBlockHashValid(), "Block hash should be invalid")
}

func TestBlock_String(t *testing.T) {

	expectedString := "0001-01-01 00:00:00 +0000 UTC\tTest\tasdf\tasdf2\t42\t4"
	block := &Block{
		Timestamp:     time.Time{},
		Data:          "Test",
		PrevBlockHash: "asdf",
		BlockHash:     "asdf2",
		Index:         42,
		Nonce:         43,
		Difficulty:    4,
	}
	t.Log(block.String())
	assert.Equal(t, expectedString, block.String(), "Block string should return a string with all fields")
}
