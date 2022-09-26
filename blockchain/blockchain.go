package blockchain

import (
	"errors"
	"github.com/defaziom/blockchain-go/block"
	"log"
	"math"
	"strings"
	"time"
)

type DoublyLinkedBlockList struct {
	Prev  *DoublyLinkedBlockList
	Next  *DoublyLinkedBlockList
	Value *block.Block
}

// Add Adds a block to the end of the list and returns the latest element
func (list *DoublyLinkedBlockList) Add(block *block.Block) *DoublyLinkedBlockList {
	list.Next = &DoublyLinkedBlockList{
		Prev:  list,
		Next:  nil,
		Value: block,
	}
	return list.Next
}

// First Returns the first element in the list
func (list *DoublyLinkedBlockList) First() *DoublyLinkedBlockList {
	if list.Prev == nil {
		return list
	}
	first := list.Prev
	for first.Prev != nil {
		first = first.Prev
	}
	return first
}

// Last Returns the element `index` from the end
func (list *DoublyLinkedBlockList) Last(index int) *DoublyLinkedBlockList {
	iter := list
	for i := 0; iter.Prev != nil && i < index; i++ {
		iter = iter.Prev
	}
	return iter
}

// ToSlice Converts the list to a slice of blocks by appending all the blocks to a slice.
func (list *DoublyLinkedBlockList) ToSlice() []*block.Block {
	slice := make([]*block.Block, 1)
	node := list.First()
	slice[0] = node.Value
	for node.Next != nil {
		node = node.Next
		slice = append(slice, node.Value)
	}
	return slice
}

func DoublyLinkedBlockListCreateFromSlice(blocks []*block.Block) *DoublyLinkedBlockList {
	newList := &DoublyLinkedBlockList{
		Value: nil,
	}
	if len(blocks) == 0 {
		return newList
	} else {
		newList.Value = blocks[0]
	}
	for _, b := range blocks[1:] {
		newList = newList.Add(b)
	}
	return newList
}

type BlockChain struct {
	Blocks *DoublyLinkedBlockList
}

var genesisBlock = &block.Block{
	Timestamp:     time.Now(),
	Data:          "Genesis",
	PrevBlockHash: "",
	BlockHash:     strings.Repeat("0", 64),
	Index:         0,
	Nonce:         0,
	Difficulty:    1,
}

const DifficultyAdjustmentIntervalBlocks = 5 // Adjusts blockchain difficulty every N blocks
const BlockGenerationIntervalSec = 0.5       // Avg interval between added blocks for adjusting difficulty

func GetGenesisBlock() *block.Block {
	return genesisBlock
}

var TheBlockChain = &BlockChain{
	Blocks: &DoublyLinkedBlockList{
		Prev:  nil,
		Next:  nil,
		Value: GetGenesisBlock(),
	},
}

// MineBlock Mines a block and returns it
func (bc *BlockChain) MineBlock(data string) *block.Block {

	lastBlock := bc.GetLatestBlock()
	b := &block.Block{
		Timestamp:     time.Now(),
		Data:          data,
		PrevBlockHash: lastBlock.BlockHash,
		BlockHash:     "",
		Index:         lastBlock.Index + 1,
		Nonce:         -1,
		Difficulty:    bc.GetDifficulty(),
	}
	blockHash := b.CalculateBlockHash()
	b.BlockHash = blockHash

	for !b.IsBlockHashValid() {
		b.Nonce += 1
		blockHash = b.CalculateBlockHash()
		b.BlockHash = blockHash
	}

	return b
}

// AddBlock Adds a block to the end of the blockchain. Check to see if the new block is valid.
func (bc *BlockChain) AddBlock(block *block.Block) error {

	lastBlock := bc.GetLatestBlock()
	valid, err := IsNewBlockValid(block, lastBlock)
	if valid {
		bc.Blocks = bc.Blocks.Add(block)
		return nil
	} else {
		return err
	}
}

func (bc *BlockChain) GetDifficulty() int {
	latestBlock := bc.GetLatestBlock()

	if latestBlock.Index%DifficultyAdjustmentIntervalBlocks == 0 && latestBlock.Index != 0 {
		return bc.GetAdjustedDifficulty()
	} else {
		return latestBlock.Difficulty
	}

}

func (bc *BlockChain) GetAdjustedDifficulty() int {
	latestBlock := bc.GetLatestBlock()
	prevAdjBlock := bc.Blocks.Last(DifficultyAdjustmentIntervalBlocks).Value
	timeExpectedSec := BlockGenerationIntervalSec * DifficultyAdjustmentIntervalBlocks
	timeTaken := latestBlock.Timestamp.Sub(prevAdjBlock.Timestamp).Seconds()

	if timeTaken < timeExpectedSec/2 {
		// Blocks are being generated too quickly, increase the difficulty
		return prevAdjBlock.Difficulty + 1
	} else if timeTaken > timeExpectedSec*2 && prevAdjBlock.Difficulty > 1 {
		// Blocks are being generated too slowly, decrease the difficulty
		return prevAdjBlock.Difficulty - 1
	} else {
		// Blocks are being generated within the tolerance, keep difficulty the same
		return prevAdjBlock.Difficulty
	}
}

func (bc *BlockChain) GetCumulativeDifficulty() float64 {
	difficultySum := 0.0
	for _, b := range bc.Blocks.ToSlice() {
		difficultySum += math.Pow(2, float64(b.Difficulty))
	}
	return difficultySum
}

func (bc *BlockChain) GetLatestBlock() *block.Block {
	return bc.Blocks.Value
}

func (bc *BlockChain) ReplaceChain(newChain *BlockChain) {
	if IsValidBlockChain(newChain) && newChain.GetCumulativeDifficulty() > bc.GetCumulativeDifficulty() {
		log.Println("Received blockchain is valid. Replacing current blockchain with received blockchain")
		bc.Blocks = newChain.Blocks
	} else {
		log.Println("Received blockchain is invalid.")
	}
}

// IsNewBlockValid Checks if a new block is valid to go on the end of the blockchain
func IsNewBlockValid(newBlock *block.Block, prevBlock *block.Block) (bool, error) {
	if newBlock.Index != prevBlock.Index+1 {
		return false, errors.New("invalid block index")
	} else if newBlock.PrevBlockHash != prevBlock.BlockHash {
		return false, errors.New("invalid prev block hash")
	} else if newBlock.BlockHash != newBlock.CalculateBlockHash() {
		return false, errors.New("invalid block hash")
	} else {
		return true, nil
	}
}

func IsValidGenesisBlock(block *block.Block) bool {
	return block == GetGenesisBlock()
}

func IsValidBlockChain(bc *BlockChain) bool {
	list := bc.Blocks
	for list.Prev != nil {
		currentBlock := list.Value
		prevBlock := list.Prev.Value
		valid, _ := IsNewBlockValid(currentBlock, prevBlock)
		if !valid {
			return false
		}
		list = list.Prev
	}
	// Last block should be genesis block
	return IsValidGenesisBlock(list.Value)
}
