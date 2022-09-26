package blockchain

import (
	"fmt"
	"github.com/defaziom/blockchain-go/block"
	"math"
	"testing"
	"time"
)
import "github.com/stretchr/testify/assert"

func CreateBlockChain() *BlockChain {
	return &BlockChain{
		Blocks: &SafeDoublyLinkedBlockList{
			Prev:  nil,
			Next:  nil,
			Value: GetGenesisBlock(),
		},
	}
}

func TestDoublyLinkedBlockList_Add(t *testing.T) {
	b1 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "first block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b2 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "second block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}

	first := &SafeDoublyLinkedBlockList{
		Prev:  nil,
		Next:  nil,
		Value: b1,
	}
	list := first.Add(b2)

	assert.Equal(t, b2, list.Value)
	assert.Equal(t, first, list.Prev)
	assert.Nil(t, list.Next)
	assert.Equal(t, list, first.Next)

}

func TestDoublyLinkedBlockList_First(t *testing.T) {
	b1 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "first block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b2 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "second block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}

	expectedFirst := &SafeDoublyLinkedBlockList{
		Prev:  nil,
		Next:  nil,
		Value: b1,
	}
	list := expectedFirst.Add(b2)

	assert.Equal(t, expectedFirst, list.First())
}

func TestDoublyLinkedBlockList_Last(t *testing.T) {

	b1 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "first block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b2 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "second block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b3 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "third block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	testList := &SafeDoublyLinkedBlockList{
		Prev:  nil,
		Next:  nil,
		Value: b1,
	}
	testList = testList.Add(b2)
	testList = testList.Add(b3)

	actualB3 := testList.Last(0).Value
	assert.Equal(t, b3, actualB3)
	actualB2 := testList.Last(1).Value
	assert.Equal(t, b2, actualB2)
	actualB1 := testList.Last(2).Value
	assert.Equal(t, b1, actualB1)
	actualB1 = testList.Last(99).Value
	assert.Equal(t, b1, actualB1)
}

func TestDoublyLinkedBlockList_ToSlice(t *testing.T) {
	b1 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "first block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b2 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "second block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}

	first := &SafeDoublyLinkedBlockList{
		Prev:  nil,
		Next:  nil,
		Value: b1,
	}
	list := first.Add(b2)
	expectedSlice := []*block.Block{b1, b2}

	assert.Equal(t, expectedSlice, list.ToSlice())

}

func TestDoublyLinkedBlockList_AddFromSlice(t *testing.T) {
	b1 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "first block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b2 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "second block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	b3 := &block.Block{
		Timestamp:     time.Now(),
		Data:          "third block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}

	expectedBlocks := []*block.Block{b1, b2, b3}

	newList := DoublyLinkedBlockListCreateFromSlice(expectedBlocks)
	actualBlocks := newList.ToSlice()

	assert.Equal(t, expectedBlocks, actualBlocks)

}

func TestBlockChain_MineBlock(t *testing.T) {
	const expectedData = "asdf"

	minedBlock := CreateBlockChain().MineBlock(expectedData)

	assert.Equal(t, expectedData, minedBlock.Data, "Block should contain input data")
	assert.NotEqualf(t, time.Time{}, minedBlock.Timestamp, "Timestamp should be initialized")
	assert.Equal(t, GetGenesisBlock().BlockHash, minedBlock.PrevBlockHash,
		"Prev block hash must equal genesis block hash")
	assert.Equal(t, "0", minedBlock.BlockHash[:minedBlock.Difficulty],
		"Block hash must be prefixed by leading zeros equal to the difficulty")
	assert.Equal(t, 1, minedBlock.Index, "Index must be 1")
}

func TestBlockChain_IsNewBlockValid(t *testing.T) {
	prevBlock := &block.Block{
		Timestamp:     time.Now(),
		Data:          "prev block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	prevBlock.BlockHash = prevBlock.CalculateBlockHash()
	newBlock := &block.Block{
		Timestamp:     time.Now(),
		Data:          "new block",
		PrevBlockHash: prevBlock.BlockHash,
		BlockHash:     "",
		Index:         1,
		Nonce:         0,
		Difficulty:    0,
	}
	newBlock.BlockHash = newBlock.CalculateBlockHash()
	valid, err := IsNewBlockValid(newBlock, prevBlock)
	assert.True(t, valid, "New block should be valid")
	assert.Nil(t, err)

	newBlock.Index = 42
	valid, err = IsNewBlockValid(newBlock, prevBlock)
	assert.False(t, valid, "New block should have invalid index")
	assert.Equal(t, "invalid block index", err.Error())
	newBlock.Index = 1

	newBlock.PrevBlockHash = "invalid hash"
	valid, err = IsNewBlockValid(newBlock, prevBlock)
	assert.False(t, valid, "New block should have invalid prev hash")
	assert.Equal(t, "invalid prev block hash", err.Error())
	newBlock.PrevBlockHash = prevBlock.BlockHash

	newBlock.BlockHash = "invalid hash"
	valid, err = IsNewBlockValid(newBlock, prevBlock)
	assert.False(t, valid, "New block should have invalid hash")
	assert.Equal(t, "invalid block hash", err.Error())
}

func TestBlockChain_AddBlock(t *testing.T) {
	blockchain := CreateBlockChain()

	newBlock := &block.Block{
		Timestamp:     time.Now(),
		Data:          "new block",
		PrevBlockHash: GetGenesisBlock().BlockHash,
		BlockHash:     "",
		Index:         1,
		Nonce:         0,
		Difficulty:    4,
	}
	newBlock.BlockHash = newBlock.CalculateBlockHash()

	err := blockchain.AddBlock(newBlock)

	assert.Nil(t, err)
	assert.Equal(t, newBlock, blockchain.GetLatestBlock(), "The new block should be the added block")
	assert.Equal(t, newBlock, blockchain.Blocks.Value, "The latest block should be the added block")
}

func TestIsValidGenesisBlock(t *testing.T) {
	assert.True(t, IsValidGenesisBlock(GetGenesisBlock()))

	b := &block.Block{
		Timestamp:     time.Now(),
		Data:          "new block",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         1,
		Nonce:         0,
		Difficulty:    4,
	}
	assert.False(t, IsValidGenesisBlock(b))

	copyGenesisBlock := *GetGenesisBlock()
	assert.False(t, IsValidGenesisBlock(&copyGenesisBlock))
}

func TestIsValidBlockChain(t *testing.T) {
	blockchain := CreateBlockChain()
	_ = blockchain.AddBlock(blockchain.MineBlock("one"))
	_ = blockchain.AddBlock(blockchain.MineBlock("two"))
	_ = blockchain.AddBlock(blockchain.MineBlock("three"))

	assert.True(t, IsValidBlockChain(blockchain))

	// Tamper with blockchain data
	blockchain.Blocks.Prev.Value.Data = "fake!"
	assert.False(t, IsValidBlockChain(blockchain))
}

func TestBlockChain_GetDifficulty(t *testing.T) {
	blockchain := CreateBlockChain()

	assert.Equal(t, GetGenesisBlock().Difficulty, blockchain.GetDifficulty())

	// Add 3 blocks
	for i := 0; i < 3; i++ {
		_ = blockchain.AddBlock(blockchain.MineBlock(fmt.Sprint(i)))
	}

	assert.Equal(t, GetGenesisBlock().Difficulty, blockchain.GetDifficulty())

	// Add 2 more blocks
	for i := 0; i < 2; i++ {
		_ = blockchain.AddBlock(blockchain.MineBlock(fmt.Sprint(i)))
	}

	assert.Equal(t, GetGenesisBlock().Difficulty+1, blockchain.GetDifficulty(),
		"Difficulty should have increased by one")
}

func TestBlockChain_GetAdjustedDifficulty(t *testing.T) {
	blockchain := CreateBlockChain()

	// Add 10 blocks
	for i := 0; i < 10; i++ {
		_ = blockchain.AddBlock(blockchain.MineBlock(fmt.Sprint(i)))
	}
	assert.Equal(t, GetGenesisBlock().Difficulty+1, blockchain.GetAdjustedDifficulty(),
		"Difficulty should have increased by one")

	// Add 5 blocks with delay
	for i := 0; i < 5; i++ {
		_ = blockchain.AddBlock(blockchain.MineBlock(fmt.Sprint(i)))
		time.Sleep(1500 * time.Millisecond)
	}
	assert.Equal(t, GetGenesisBlock().Difficulty, blockchain.GetAdjustedDifficulty(),
		"Difficulty should have decreased by one")

	// Add 5 blocks with smaller delay
	for i := 0; i < 5; i++ {
		_ = blockchain.AddBlock(blockchain.MineBlock(fmt.Sprint(i)))
		time.Sleep(1000 * time.Millisecond)
	}
	assert.Equal(t, GetGenesisBlock().Difficulty, blockchain.GetAdjustedDifficulty(),
		"Difficulty should have stayed the same")
}

func TestBlockChain_GetCumulativeDifficulty(t *testing.T) {
	blockchain := CreateBlockChain()

	b1 := &block.Block{Difficulty: 1}
	b2 := &block.Block{Difficulty: 2}
	b3 := &block.Block{Difficulty: 3}

	blockchain.Blocks = blockchain.Blocks.Add(b1)
	blockchain.Blocks = blockchain.Blocks.Add(b2)
	blockchain.Blocks = blockchain.Blocks.Add(b3)

	expectedCumulativeDifficulty := math.Pow(2, float64(GetGenesisBlock().Difficulty)) + math.Pow(2, 1) + math.Pow(2,
		2) + math.Pow(2, 3)

	assert.Equal(t, expectedCumulativeDifficulty, blockchain.GetCumulativeDifficulty())
}
