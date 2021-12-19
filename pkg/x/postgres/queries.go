package postgres

import (
	"context"
	"fmt"
)

func (p *PostgresProcessor) LastProcessedBlock(ctx context.Context, chainID string) (int64, error) {
	res, err := p.conn.Query(ctx, fmt.Sprintf(lastProcessedBlockQuery, chainID))
	if err != nil {
		return -1, err
	}

	defer res.Close()

	if res.Next() {
		block := new(int64)
		err = res.Scan(block)
		if err != nil {
			return -1, err
		}
		return *block, nil
	}
	return 0, nil
}

func (p *PostgresProcessor) ChainIDFromClientID(ctx context.Context, clientID, originChainID string) (string, error) {
	res, err := p.conn.Query(ctx, fmt.Sprintf(chainIDFromClientIDQuery, clientID, originChainID))
	if err != nil {
		return "", err
	}

	defer res.Close()

	if res.Next() {
		chainID := ""
		err = res.Scan(&chainID)
		if err != nil {
			return "", err
		}
		return chainID, nil
	}
	return "", nil
}

func (p *PostgresProcessor) ChainIDFromConnectionID(ctx context.Context, connectionID, originChainID string) (string, error) {
	res, err := p.conn.Query(ctx, fmt.Sprintf(clientIDFromConnectionIDQuery, connectionID, originChainID))
	if err != nil {
		return "", err
	}
	defer res.Close()

	if res.Next() {
		clientID := ""
		err = res.Scan(&clientID)
		if err != nil {
			return "", err
		}
		res.Close()
		return p.ChainIDFromClientID(ctx, clientID, originChainID)
	}
	return "", nil
}

func (p *PostgresProcessor) ChainIDFromChannelID(ctx context.Context, channelID, originChainID string) (string, error) {
	res, err := p.conn.Query(ctx, fmt.Sprintf(connectionIDFromChannelIDQuery, channelID, originChainID))
	if err != nil {
		return "", err
	}

	defer res.Close()

	if res.Next() {
		connectionID := ""
		err = res.Scan(&connectionID)
		if err != nil {
			return "", err
		}
		res.Close()
		return p.ChainIDFromConnectionID(ctx, connectionID, originChainID)
	}
	return "", nil
}

func (p *PostgresProcessor) GetChannelStatus(ctx context.Context, channelID, originChainID string) (bool, error) {
	res, err := p.conn.Query(ctx, fmt.Sprintf(statusFromChannelIDQuery, channelID, originChainID))
	if err != nil {
		return false, err
	}

	defer res.Close()

	if res.Next() {
		isOpened := false
		err = res.Scan(&isOpened)
		if err != nil {
			return false, err
		}
		res.Close()
		return isOpened, nil
	}
	return false, nil
}

// ChainID method returns chain ID related to the given channel_id
// it checks for local(block) data and does appropriate db queries
func (p *PostgresProcessor) ChainID(ctx context.Context, channelID, originChainID string) (string, error) {
	// check block cache before attempting to query db
	// if whole chain of events(client -> connection -> channel) happened in the same block
	//channelMap := p.clients[p.connections[p.channels[channelID]]]
	channelMap := p.channels[channelID]
	var connectionId string
	for k, _ := range channelMap {
		connectionId = k
		break
	}
	if chainID, ok := p.clients[p.connections[connectionId]]; ok {
		return chainID, nil
	}
	// if connection and channel happened in this block
	if clientID, ok := p.connections[connectionId]; ok {
		return p.ChainIDFromClientID(ctx, clientID, originChainID)
	}
	// if channel was created in the same block
	if connectionMap, ok := p.channels[channelID]; ok {
		for connectionID, _ := range connectionMap {
			return p.ChainIDFromConnectionID(ctx, connectionID, originChainID)
		}
	}

	// nothing in cache, query db
	return p.ChainIDFromChannelID(ctx, channelID, originChainID)
}
