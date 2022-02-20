package processor

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestIbcData_Append(t *testing.T) {
	type args struct {
		source      string
		destination string
		channel     string
		t           time.Time
		coins       []struct {
			Amount *big.Int
			Coin   string
		}
		isFailed bool
	}
	timeArgs, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	timeWant, _ := time.Parse("2006-01-02T15:00:00", "2006-01-02T15:00:00")
	m := IbcData{}
	sourceName := "mySource"
	destinationName1 := "myDestination"
	destinationName2 := "myDestination2"
	channelID := "channel-1"
	firstAmount, _ := new(big.Int).SetString("93458345", 10)
	secondAmount, _ := new(big.Int).SetString("2345432435", 10)
	thirdAmount, _ := new(big.Int).SetString("1", 10)
	coins := []struct {
		Amount *big.Int
		Coin   string
	}{
		{
			Amount: firstAmount,
			Coin:   "ibc/sdfjlksadflkdsafkdsfj34285udfaj",
		},
		{
			Amount: secondAmount,
			Coin:   "ibc/j934u5edjf9d8fu984uteh8hfedw9fh9",
		},
		{
			Amount: thirdAmount,
			Coin:   "ibc/sdfjlksadflkdsafkdsfj34285udfaj",
		},
	}
	coinsMap := make(map[string]*big.Int)
	coinsMap["ibc/sdfjlksadflkdsafkdsfj34285udfaj"], _ = new(big.Int).SetString("93458346", 10)
	coinsMap["ibc/j934u5edjf9d8fu984uteh8hfedw9fh9"], _ = new(big.Int).SetString("2345432435", 10)
	coinsMap2 := make(map[string]*big.Int)
	coinsMap2["ibc/sdfjlksadflkdsafkdsfj34285udfaj"], _ = new(big.Int).SetString("186916692", 10)
	coinsMap2["ibc/j934u5edjf9d8fu984uteh8hfedw9fh9"], _ = new(big.Int).SetString("4690864870", 10)
	failedTx := true
	notFailedTx := false
	tests := []struct {
		name    string
		ibcData IbcData
		args    args
		want    IbcData
	}{
		{
			"initial_increment",
			m,
			args{sourceName, destinationName1, channelID, timeArgs, coins, notFailedTx},
			map[string]map[string]map[string]map[time.Time]*IbcCounters{sourceName: {destinationName1: {channelID: {timeWant: &IbcCounters{
				Transfers:       1,
				FailedTransfers: 0,
				Coin:            coinsMap,
			}}}}},
		},
		{
			"initial_increment_without_cashflow",
			IbcData{},
			args{sourceName, destinationName1, channelID, timeArgs, coins, failedTx},
			map[string]map[string]map[string]map[time.Time]*IbcCounters{sourceName: {destinationName1: {channelID: {timeWant: &IbcCounters{
				Transfers:       1,
				FailedTransfers: 1,
				Coin:            nil,
			}}}}},
		},
		{
			"increment_existing",
			m,
			args{sourceName, destinationName1, channelID, timeArgs, coins, notFailedTx},
			map[string]map[string]map[string]map[time.Time]*IbcCounters{sourceName: {destinationName1: {channelID: {timeWant: &IbcCounters{
				Transfers:       2,
				FailedTransfers: 0,
				Coin:            coinsMap2,
			}}}}},
		},
		{
			"increment_with_second_destination",
			m,
			args{sourceName, destinationName2, channelID, timeArgs, coins, failedTx},
			map[string]map[string]map[string]map[time.Time]*IbcCounters{sourceName: {destinationName1: {channelID: {timeWant: &IbcCounters{
				Transfers:       2,
				FailedTransfers: 0,
				Coin:            coinsMap2,
			}}}, destinationName2: {channelID: {timeWant: &IbcCounters{
				Transfers:       1,
				FailedTransfers: 1,
				Coin:            nil,
			}}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ibcData.Append(tt.args.source, tt.args.destination, tt.args.t, tt.args.channel, tt.args.isFailed, coins)
			assert.Equal(t, tt.want, tt.ibcData)
		})
	}
}

func TestIbcData_ToIbcStats(t *testing.T) {
	timeArgs, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	sourceName := "mySource"
	destinationName1 := "myDestination"
	destinationName2 := "myDestination2"
	channelID := "channel-1"
	transferCounter1 := 2
	transferCounter2 := 7
	failedTransferCounter1 := 1
	failedTransferCounter2 := 3
	tests := []struct {
		name     string
		ibcData  IbcData
		expected [][]IbcStats
	}{
		{
			"IbcData(map)_to_IbcStats(slice)",
			map[string]map[string]map[string]map[time.Time]*IbcCounters{sourceName: {destinationName1: {channelID: {timeArgs: &IbcCounters{
				Transfers:       transferCounter1,
				FailedTransfers: failedTransferCounter1,
			}}}, destinationName2: {channelID: {timeArgs: &IbcCounters{
				Transfers:       transferCounter2,
				FailedTransfers: failedTransferCounter2,
			}}}}},
			[][]IbcStats{
				{
					{sourceName, destinationName1, channelID, timeArgs, transferCounter1, failedTransferCounter1},
					{sourceName, destinationName2, channelID, timeArgs, transferCounter2, failedTransferCounter2},
				},
				{
					{sourceName, destinationName2, channelID, timeArgs, transferCounter2, failedTransferCounter2},
					{sourceName, destinationName1, channelID, timeArgs, transferCounter1, failedTransferCounter1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.ibcData.ToIbcStats()

			if !reflect.DeepEqual(tt.expected[0], actual) {
				assert.Equal(t, tt.expected[1], actual)
			} else {
				assert.NotEqual(t, tt.expected[1], actual)
			}

			if !reflect.DeepEqual(tt.expected[1], actual) {
				assert.Equal(t, tt.expected[0], actual)
			} else {
				assert.NotEqual(t, tt.expected[0], actual)
			}
		})
	}
}
