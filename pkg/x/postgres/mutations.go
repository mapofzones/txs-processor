package postgres

import (
	"fmt"
	"time"

	processor "github.com/mapofzones/txs-processor/pkg/types"
)

func addZone(chainID string) string {
	return fmt.Sprintf(addZoneQuery,
		fmt.Sprintf("('%s', '%s', %t, %t)", chainID, chainID, true, false),
		true,
	)
}

func addImplicitZones(clients map[string]string) string {
	query := ""
	for _, chainID := range clients {
		if len(chainID) > 0 {
			query += fmt.Sprintf("('%s', '%s', %t, %t),", chainID, chainID, false, false)
		}
	}
	result := ""
	if len(query) > 0 {
		query = query[:len(query)-1]
		result = fmt.Sprintf(addImplicitZoneQuery, query)
	}
	return result
}

func markBlock(chainID string) string {
	t := time.Now().Format(Format)
	return markBlockConstruct(chainID, t)
}

func markBlockConstruct(chainID string, t string) string {
	return fmt.Sprintf(markBlockQuery,
		fmt.Sprintf("('%s', %d, '%s')", chainID, 1, t), t)
}

func addTxStats(stats processor.TxStats) string {
	return fmt.Sprintf(addTxStatsQuery,
		fmt.Sprintf("('%s', '%s', %d, %d, %d, %d, %d)", stats.ChainID, stats.Hour.Format(Format), stats.Count,
			stats.TxWithIBCTransfer, 1, stats.TxWithIBCTransferFail, stats.TurnoverAmount),
		stats.Count,
		stats.TxWithIBCTransfer,
		stats.TxWithIBCTransferFail,
		stats.TurnoverAmount,
	)
}

func addActiveAddressesStats(stats processor.TxStats, addressData processor.AddressData) string {
	return fmt.Sprintf(addActiveAddressesQuery,
		fmt.Sprintf(
			"('%s', '%s', '%s', %d, %t, %t, %t)",
			addressData.Address,
			stats.ChainID,
			stats.Hour.Format(Format),
			1,
			addressData.IsInternalTx,
			addressData.IsInternalTransfer,
			addressData.IsExternalTransfer,
		),
	)
}

func addClients(origin string, clients map[string]string) string {
	values := ""
	for clientID, chainID := range clients {
		var row string
		if len(chainID) > 0 {
			row = fmt.Sprintf("('%s', '%s', '%s'),", origin, clientID, chainID)
		} else {
			row = fmt.Sprintf("('%s', '%s', null),", origin, clientID)
		}
		values += row
	}
	if len(values) > 0 {
		values = values[:len(values)-1]
	}
	return fmt.Sprintf(addClientsQuery, values)
}

func addConnections(origin string, data map[string]string) string {
	values := ""
	for connectionID, clientID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, connectionID, clientID)
	}
	if len(values) > 0 {
		values = values[:len(values)-1]
	}

	return fmt.Sprintf(addConnectionsQuery, values)
}

func addChannels(origin string, data map[string]string) string {
	values := ""
	for channelID, connectionID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s',%t),", origin, channelID, connectionID, false)
	}
	if len(values) > 0 {
		values = values[:len(values)-1]
	}

	return fmt.Sprintf(addChannelsQuery, values)
}

func markChannel(origin, channelID string, state bool) string {
	return fmt.Sprintf(markChannelQuery,
		state,
		origin,
		channelID)
}

func addIbcStats(origin string, ibcData map[string]map[string]map[string]map[time.Time]*processor.IbcCounters) []string {
	// buffer for our queries
	queries := make([]string, 0, 32)

	// process ibc transfers
	for source, destMap := range ibcData {
		for dest, hourMap := range destMap {
			for channel, channelMap := range hourMap {
				for hour, count := range channelMap {
					queries = append(queries, fmt.Sprintf(addIbcStatsQuery,
						fmt.Sprintf("('%s', '%s', '%s', '%s', %d, %d, '%s', %d)", origin, source, dest, hour.Format(Format), count.Transfers, 1, channel, count.FailedTransfers),
						count.Transfers, count.FailedTransfers))
					for denom, amount := range count.Coin {
						queries = append(queries, fmt.Sprintf(addIbcCashflowQuery,
							fmt.Sprintf("('%s', '%s', '%s', '%s', %d, '%s', '%s', %d)", origin, source, dest, hour.Format(Format), 1, channel, denom, amount),
							amount))
					}
				}
			}
		}
	}
	return queries
}
