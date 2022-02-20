package postgres

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	processor "github.com/mapofzones/txs-processor/pkg/types"
)

func (p *PostgresProcessor) handleTransaction(ctx context.Context, metadata processor.MessageMetadata, msg watcher.Transaction) error {
	// this should not happen
	if metadata.TxMetadata == nil {
		panic(fmt.Errorf("%w: could not fetch tx metadata", processor.CommitError))
	}

	if p.txStats == nil {
		p.txStats = &processor.TxStats{
			ChainID:        metadata.ChainID,
			Hour:           metadata.BlockTime.Truncate(time.Hour),
			TurnoverAmount: big.NewInt(0),
		}
	}

	// addresses collection logic
	if len(msg.Sender) > 0 {
		address := &processor.AddressData{
			Address:            msg.Sender,
			IsInternalTx:       true,
			IsInternalTransfer: false,
			IsExternalTransfer: false,
		}
		p.txStats.Addresses = append(p.txStats.Addresses, address)
	} else {
		log.Println("Not found sender for tx!")
	}

	// if tx had errors and did not affect the state
	if !metadata.TxMetadata.Accepted {
		p.txStats.Count++
		for _, m := range msg.Messages {
			if message, ok := m.(watcher.IBCTransfer); ok {
				p.txStats.TxWithIBCTransferFail++
				p.txStats.TxWithIBCTransfer++
				p.handleIBCTransfer(ctx, metadata, message)
				return nil
			}
		}
		return nil
	}

	if p.txStats.TurnoverAmount == nil {
		p.txStats.TurnoverAmount = big.NewInt(0)
	}

	hasIBCTransfers := false
	// process each tx message
	for _, m := range msg.Messages {
		if _, ok := m.(watcher.IBCTransfer); ok {
			hasIBCTransfers = true
			for _, am := range m.(watcher.IBCTransfer).Amount {
				p.txStats.TurnoverAmount.Add(p.txStats.TurnoverAmount, am.Amount)
			}
			address := &processor.AddressData{
				Address:            m.(watcher.IBCTransfer).Sender,
				IsInternalTx:       false,
				IsInternalTransfer: m.(watcher.IBCTransfer).Source,
				IsExternalTransfer: !m.(watcher.IBCTransfer).Source,
			}
			p.txStats.Addresses = append(p.txStats.Addresses, address)
			log.Println(m.(watcher.IBCTransfer).Sender)
		}
		if _, ok := m.(watcher.Transfer); ok {
			for _, am := range m.(watcher.Transfer).Amount {
				p.txStats.TurnoverAmount.Add(p.txStats.TurnoverAmount, am.Amount)
			}
			address := &processor.AddressData{
				Address:            m.(watcher.Transfer).Sender,
				IsInternalTx:       true,
				IsInternalTransfer: false,
				IsExternalTransfer: false,
			}
			p.txStats.Addresses = append(p.txStats.Addresses, address)
			log.Println(m.(watcher.Transfer).Sender)
		}
		handle := p.Handler(m)
		if handle != nil {
			err := handle(ctx, metadata, m)
			if err != nil {
				return err
			}
		}
	}

	// increment tx stats
	p.txStats.Count++
	// if tx had ibc transfers, mark it
	if hasIBCTransfers {
		p.txStats.TxWithIBCTransfer++
	}

	return nil
}

func (p *PostgresProcessor) handleCreateClient(ctx context.Context, metadata processor.MessageMetadata, msg watcher.CreateClient) error {
	p.clients[msg.ClientID] = msg.ChainID
	return nil
}

func (p *PostgresProcessor) handleCreateConnection(ctx context.Context, metadata processor.MessageMetadata, msg watcher.CreateConnection) error {
	p.connections[msg.ConnectionID] = msg.ClientID
	return nil
}

func (p *PostgresProcessor) handleCreateChannel(ctx context.Context, metadata processor.MessageMetadata, msg watcher.CreateChannel) error {
	p.channels[msg.ChannelID] = msg.ConnectionID
	return nil
}

func (p *PostgresProcessor) handleOpenChannel(ctx context.Context, metadata processor.MessageMetadata, msg watcher.OpenChannel) error {
	p.channelStates[msg.ChannelID] = true
	return nil
}

func (p *PostgresProcessor) handleCloseChannel(ctx context.Context, metadata processor.MessageMetadata, msg watcher.CloseChannel) error {
	p.channelStates[msg.ChannelID] = false
	return nil
}

func (p *PostgresProcessor) handleIBCTransfer(ctx context.Context, metadata processor.MessageMetadata, msg watcher.IBCTransfer) error {
	chainID, err := p.ChainID(ctx, msg.ChannelID, metadata.ChainID)
	if err != nil {
		return fmt.Errorf("%w: %s", processor.ConnectionError, err.Error())
	}
	if chainID == "" {
		//todo: need to enable Errorf when will be solomachine support added
		//return fmt.Errorf("%w: could not fetch chainID connected to given channelID", processor.CommitError)
		return nil
	}

	isEnabledChannel, err := p.GetChannelStatus(ctx, msg.ChannelID, metadata.ChainID)
	if err != nil || isEnabledChannel == false {
		//return fmt.Errorf("%w: could not process ibc transfer with closed channelID", processor.CommitError)
		//todo: need to recalculate statistics for frozen transfer txs and resolve the issue of transactions to closed channels
	}

	if len(chainID) > 0 { // ignore solomachines todo: need to add solomachines support
		if msg.Source {
			p.ibcStats.Append(metadata.ChainID, chainID, metadata.BlockTime, msg.ChannelID, !metadata.TxMetadata.Accepted, msg.Amount)
		} else {
			p.ibcStats.Append(chainID, metadata.ChainID, metadata.BlockTime, msg.ChannelID, !metadata.TxMetadata.Accepted, msg.Amount)
		}
	}

	return nil
}
