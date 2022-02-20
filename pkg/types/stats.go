package processor

import (
	"math/big"
	"time"
)

type AddressData struct {
	Address            string
	IsInternalTx       bool
	IsInternalTransfer bool
	IsExternalTransfer bool
}

// TxStats structure is used to see how many txs were send during each hour
type TxStats struct {
	ChainID               string
	Hour                  time.Time //must have 0 minutes, seconds and micro/nano seconds
	Count                 int
	TxWithIBCTransfer     int
	TxWithIBCTransferFail int
	Addresses             []*AddressData
	TurnoverAmount        *big.Int
}

// IbcStats represents statistics that we need to write to db
type IbcStats struct {
	Source      string
	Destination string
	Channel     string
	Hour        time.Time //must have 0 minutes, seconds and micro/nano seconds
	Count       int
	FailedCount int
}

type IbcCounters struct {
	Transfers       int
	FailedTransfers int
	Coin            map[string]*big.Int
}

// IbcData is used to organize ibc tx data during each hour
type IbcData map[string]map[string]map[string]map[time.Time]*IbcCounters

// Append truncates timestamps and puts data into ibc data structure
func (m *IbcData) Append(source, destination string, t time.Time, channelID string, isFailed bool, coins []struct {
	Amount *big.Int
	Coin   string
}) {
	t = t.Truncate(time.Hour)
	if *m == nil {
		*m = make(IbcData)
	}

	if (*m)[source] == nil {
		(*m)[source] = make(map[string]map[string]map[time.Time]*IbcCounters)
	}

	if (*m)[source][destination] == nil {
		(*m)[source][destination] = make(map[string]map[time.Time]*IbcCounters)
	}

	if (*m)[source][destination][channelID] == nil {
		(*m)[source][destination][channelID] = make(map[time.Time]*IbcCounters)
	}

	if (*m)[source][destination][channelID][t] == nil {
		(*m)[source][destination][channelID][t] = &IbcCounters{
			Transfers:       0,
			FailedTransfers: 0,
			Coin:            nil,
		}
	}

	((*m)[source][destination][channelID][t]).Transfers++
	if isFailed {
		((*m)[source][destination][channelID][t]).FailedTransfers++
	} else {
		if ((*m)[source][destination][channelID][t]).Coin == nil {
			((*m)[source][destination][channelID][t]).Coin = make(map[string]*big.Int)
		}

		for _, coin := range coins {
			value := ((*m)[source][destination][channelID][t]).Coin[coin.Coin]
			if value == nil {
				value = new(big.Int)
				((*m)[source][destination][channelID][t]).Coin[coin.Coin] = value
			}
			value.Add(value, coin.Amount)
		}
	}
}

// ToIbcStats returns slice of ibc stats formed from ibcData maps
func (m IbcData) ToIbcStats() []IbcStats {
	var stats []IbcStats
	for source := range m {
		for destination := range m[source] {
			for channel := range m[source][destination] {
				for hour := range m[source][destination][channel] {
					stats = append(stats, IbcStats{
						Source:      source,
						Destination: destination,
						Channel:     channel,
						Hour:        hour,
						Count:       m[source][destination][channel][hour].Transfers,
						FailedCount: m[source][destination][channel][hour].FailedTransfers,
					})
				}
			}
		}
	}
	return stats
}
