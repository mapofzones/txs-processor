package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	processor "github.com/mapofzones/txs-processor/pkg/types"
)

// compile time check
var _ processor.Processor = &PostgresProcessor{}

type PostgresProcessor struct {
	conn          *pgx.Conn
	txStats       *processor.TxStats
	ibcStats      processor.IbcData
	clients       map[string]string
	connections   map[string]string
	channels      map[string]string
	channelStates map[string]bool
}

// NewProcessor returns instance of Postgres processor
func NewProcessor(ctx context.Context, dbEndpoint string) (*PostgresProcessor, error) {
	conn, err := pgx.Connect(ctx, dbEndpoint)
	if err != nil {
		return nil, err
	}
	return &PostgresProcessor{
		conn:          conn,
		clients:       make(map[string]string),
		connections:   make(map[string]string),
		channels:      make(map[string]string),
		channelStates: make(map[string]bool),
		txStats:       nil,
		ibcStats:      nil,
	}, nil
}

// Validate checks if the block that we received is at valid height
func (p *PostgresProcessor) Validate(ctx context.Context, b watcher.Block) error {
	dbHeight, err := p.LastProcessedBlock(ctx, b.ChainID())
	// something is wrong with our database connection/query
	if err != nil {
		return fmt.Errorf("%w: %s", processor.ConnectionError, err)
	}
	// received block at wrong height
	if b.Height()-dbHeight != 1 {
		return fmt.Errorf("%w: expected block at height %d, got block at height %d", processor.BlockHeightError, dbHeight+1, b.Height())
	}
	log.Println(b.ChainID(), "\t", b.Height())
	return nil
}

func (p *PostgresProcessor) Handler(watcher.Message) func(context.Context, processor.MessageMetadata, watcher.Message) error {
	return func(ctx context.Context, metadata processor.MessageMetadata, msg watcher.Message) error {

		switch msg := msg.(type) {
		case watcher.Transaction:
			metadata.AddTxMetadata(msg)
			return p.handleTransaction(ctx, metadata, msg)

		case watcher.CreateClient:
			return p.handleCreateClient(ctx, metadata, msg)

		case watcher.CreateConnection:
			return p.handleCreateConnection(ctx, metadata, msg)

		case watcher.CreateChannel:
			return p.handleCreateChannel(ctx, metadata, msg)

		case watcher.OpenChannel:
			return p.handleOpenChannel(ctx, metadata, msg)

		case watcher.CloseChannel:
			return p.handleCloseChannel(ctx, metadata, msg)

		case watcher.IBCTransfer:
			return p.handleIBCTransfer(ctx, metadata, msg)

		default:
			return nil
		}

	}
}

func (p *PostgresProcessor) reset() {
	p.txStats = nil
	p.ibcStats = nil
	p.clients = make(map[string]string)
	p.connections = make(map[string]string)
	p.channels = make(map[string]string)
	p.channelStates = make(map[string]bool)
}

func (p *PostgresProcessor) Commit(ctx context.Context, block watcher.Block) error {
	// clear data gathered during block parsing
	// after we commit block data to db
	defer p.reset()

	batch := &pgx.Batch{}

	// add zone
	batch.Queue((addZone(block.ChainID())))

	// mark block as processed
	batch.Queue(markBlock(block.ChainID()))

	// update TxStats
	if p.txStats != nil {
		batch.Queue(addTxStats(*p.txStats))
		for _, address := range p.txStats.Addresses {
			batch.Queue(addActiveAddressesStats(*p.txStats, address))
		}
	}

	// insert ibc clients
	if len(p.clients) > 0 {
		// add zones to which clients refer
		batch.Queue(addImplicitZones(p.clients))
		// now we can add clients
		batch.Queue(addClients(block.ChainID(), p.clients))
	}

	// insert ibc connections
	if len(p.connections) > 0 {
		batch.Queue(addConnections(block.ChainID(), p.connections))
	}

	// insert ibc channels
	if len(p.channels) > 0 {
		batch.Queue(addChannels(block.ChainID(), p.channels))
	}

	// update channelStates
	for channel, state := range p.channelStates {
		batch.Queue(markChannel(block.ChainID(), channel, state))
	}

	// update ibc stats and add untraced zones
	for _, query := range addIbcStats(block.ChainID(), p.ibcStats) {
		batch.Queue(query)
	}

	res := p.conn.SendBatch(ctx, batch)
	defer res.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := res.Exec()
		if err != nil {
			return fmt.Errorf("%w: %s", processor.CommitError, err.Error())
		}
	}
	log.Println("chain_id: ", block.ChainID(), " height: ", block.Height())
	return nil
}
