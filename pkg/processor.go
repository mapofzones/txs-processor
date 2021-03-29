package processor

import (
	"context"
	"log"

	"errors"

	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	processor "github.com/mapofzones/txs-processor/pkg/types"
)

// Processor holds handles for all our connections
// and holds an interface which defines what has to be done
// on the received block
type Processor struct {
	Blocks <-chan watcher.Block
	processor.Processor
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, blocks <-chan watcher.Block, blockProcessor processor.Processor) *Processor {
	return &Processor{
		Blocks:    blocks,
		Processor: blockProcessor,
	}
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// this map is used to avoid constant spam of invalid height messages if that
	// error occurs
	ignoredChains := map[string]bool{}

	// receive from block stream and process
	for {
		select {
		case block, ok := <-p.Blocks:
			if !ok {
				return errors.New("block channel is closed")
			}

			if err := p.ProcessBlock(ctx, block); err != nil {
				// if we have error in our logic or there is no connection
				if errors.Is(err, processor.ConnectionError) ||
					errors.Is(err, processor.CommitError) {
					return err
				}

				// queue was fixed, no need to suppress messagess from it anymore
				if _, ok := ignoredChains[block.ChainID()]; ok && err == nil {
					delete(ignoredChains, block.ChainID())
				}

				// log the error if we are not ignoring this chain
				if _, ok := ignoredChains[block.ChainID()]; !ok {
					log.Printf("could not process block from %s: %s\n", block.ChainID(), err)
				}

				// if order of blocks is messed up, ignore it until queue is fixed
				if errors.Is(err, processor.BlockHeightError) {
					ignoredChains[block.ChainID()] = true
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) ProcessBlock(ctx context.Context, block watcher.Block) error {
	log.Println("start process: chain_id: ", block.ChainID(), " height: ", block.Height())
	err := p.Validate(ctx, block)
	if err != nil {
		return err
	}

	for _, message := range block.Messages() {
		handler := p.Handler(message)
		if handler != nil {
			err := handler(ctx, processor.MessageMetadata{
				ChainID:   block.ChainID(),
				BlockTime: block.Time(),
			}, message)
			if err != nil {
				return err
			}
		}
	}

	return p.Commit(ctx, block)
}
