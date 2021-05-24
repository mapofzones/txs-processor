package postgres

import (
    processor "github.com/mapofzones/txs-processor/pkg/types"
    "github.com/stretchr/testify/assert"
    "math/big"
    "testing"
    "time"
)

func Test_addZone(t *testing.T) {
    type args struct {
        chainID string
    }
    tests := []struct {
        name        string
        args        args
        expected    string
    }{
        {"empty_args", args{}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('', '', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
        {"first_args", args{"myChain1"}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('myChain1', 'myChain1', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
        {"second_args", args{"myChain2"}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('myChain2', 'myChain2', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addZone(tt.args.chainID)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addImplicitZones(t *testing.T) {
    type args struct {
        clients map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {"empty_args", args{}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values \n    on conflict (chain_id) do nothing;"},
        {"first_pair", args{map[string]string{"clientId1":"chainId1"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId1', 'chainId1', false, false)\n    on conflict (chain_id) do nothing;"},
        {"second_pair", args{map[string]string{"clientId2":"chainId2"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId2', 'chainId2', false, false)\n    on conflict (chain_id) do nothing;"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addImplicitZones(tt.args.clients)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_markBlockConstruct(t *testing.T) {
    type args struct {
        chainID string
        time    string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {"empty_args", args{}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('', 1, '')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '';"},
        {"first_args", args{"chainID1", "2006-01-02T15:04:05"}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('chainID1', 1, '2006-01-02T15:04:05')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '2006-01-02T15:04:05';"},
        {"second_args", args{"chainID2", "2016-12-02T06:14:55"}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('chainID2', 1, '2016-12-02T06:14:55')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '2016-12-02T06:14:55';"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := markBlockConstruct(tt.args.chainID, tt.args.time)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addTxStats(t *testing.T) {
    type args struct {
        stats processor.TxStats
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 0, 0, 1, 0, <nil>)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 0,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 0,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 0,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + <nil>;",
        },
        {
            "first_args",
            args{processor.TxStats{Count: 1, TxWithIBCTransfer: 2, TxWithIBCTransferFail: 3, TurnoverAmount: big.NewInt(11111122222333333)}},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 1, 2, 1, 3, 11111122222333333)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 1,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 2,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 3,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + 11111122222333333;",
        },
        {
            "second_args",
            args{processor.TxStats{Count: 348, TxWithIBCTransfer: 3952, TxWithIBCTransferFail: 842, TurnoverAmount: big.NewInt(877581450957345)}},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 348, 3952, 1, 842, 877581450957345)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 348,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 3952,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 842,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + 877581450957345;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addTxStats(tt.args.stats)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addActiveAddressesStats(t *testing.T) {
    timeArgs, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
    timeArgs2, _ := time.Parse("2006-01-02T15:04:05", "2018-13-11T09:17:22")
    type args struct {
        stats   processor.TxStats
        address string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "first_empty_args",
            args{},
            "insert into active_addresses(address, zone, hour, period) values ('', '', '0001-01-01T00:00:00', 1)\n    on conflict (address, zone, hour, period) do nothing;",
        },
        {
            "second_empty_args",
            args{processor.TxStats{}, ""},
            "insert into active_addresses(address, zone, hour, period) values ('', '', '0001-01-01T00:00:00', 1)\n    on conflict (address, zone, hour, period) do nothing;",
        },
        {
            "first_args",
            args{processor.TxStats{ChainID: "myChainID", Hour: timeArgs}, "moz:kfdjf928hfjvnczmnvsohvyuqoefiudb"},
            "insert into active_addresses(address, zone, hour, period) values ('moz:kfdjf928hfjvnczmnvsohvyuqoefiudb', 'myChainID', '2006-01-02T15:04:05', 1)\n    on conflict (address, zone, hour, period) do nothing;",
        },
        {
            "second_args",
            args{processor.TxStats{ChainID: "myChainID2", Hour: timeArgs2}, "moz:df89hrui3kjdf8iydhgayud"},
            "insert into active_addresses(address, zone, hour, period) values ('moz:df89hrui3kjdf8iydhgayud', 'myChainID2', '0001-01-01T00:00:00', 1)\n    on conflict (address, zone, hour, period) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addActiveAddressesStats(tt.args.stats, tt.args.address)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addClients(t *testing.T) {
    type args struct {
        origin  string
        clients map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_clients(zone, client_id, chain_id) values \n    on conflict (zone, client_id) do nothing;",
        },
        {
            "first_args",
            args{"myOrigin1", map[string]string{"clientID1":"chainID1"}},
            "insert into ibc_clients(zone, client_id, chain_id) values ('myOrigin1', 'clientID1', 'chainID1')\n    on conflict (zone, client_id) do nothing;",
        },
        {
            "second_args",
            args{"myOrigin2", map[string]string{"clientID2":"chainID2"}},
            "insert into ibc_clients(zone, client_id, chain_id) values ('myOrigin2', 'clientID2', 'chainID2')\n    on conflict (zone, client_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addClients(tt.args.origin, tt.args.clients)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addConnections(t *testing.T) {
    type args struct {
        origin string
        data   map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_connections(zone, connection_id, client_id) values \n    on conflict (zone, connection_id) do nothing;",
        },
        {
            "first_args",
            args{"origin1", map[string]string{"connectionID1": "clientID1"}},
            "insert into ibc_connections(zone, connection_id, client_id) values ('origin1', 'connectionID1', 'clientID1')\n    on conflict (zone, connection_id) do nothing;",
        },
        {
            "second_args",
            args{"origin2", map[string]string{"connectionID2": "clientID2"}},
            "insert into ibc_connections(zone, connection_id, client_id) values ('origin2', 'connectionID2', 'clientID2')\n    on conflict (zone, connection_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addConnections(tt.args.origin, tt.args.data)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addChannels(t *testing.T) {
    type args struct {
        origin string
        data   map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values \n    on conflict(zone, channel_id) do nothing;",
        },
        {
            "first_args",
            args{"origin1", map[string]string{"channelID1": "connectionID1"}},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values ('origin1', 'channelID1', 'connectionID1',false)\n    on conflict(zone, channel_id) do nothing;",
        },
        {
            "second_args",
            args{"origin2", map[string]string{"channelID2": "connectionID2"}},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values ('origin2', 'channelID2', 'connectionID2',false)\n    on conflict(zone, channel_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addChannels(tt.args.origin, tt.args.data)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_markChannel(t *testing.T) {
    type args struct {
        origin    string
        channelID string
        state     bool
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "update ibc_channels\n    set is_opened = false\n        where zone = ''\n        and channel_id = '';",
        },
        {
            "first_args",
            args{"origin1", "myChannelID1", true},
            "update ibc_channels\n    set is_opened = true\n        where zone = 'origin1'\n        and channel_id = 'myChannelID1';",
        },
        {
           "second_args",
           args{"origin2", "myChannelID2", false},
            "update ibc_channels\n    set is_opened = false\n        where zone = 'origin2'\n        and channel_id = 'myChannelID2';",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := markChannel(tt.args.origin, tt.args.channelID, tt.args.state)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addIbcStats(t *testing.T) {
    timeArgs1, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
    timeArgs2, _ := time.Parse("2006-01-02T15:04:05", "2017-09-11T04:20:49")
    type args struct {
        origin  string
        ibcData map[string]map[string]map[string]map[time.Time]int
    }
    tests := []struct {
        name string
        args args
        expected []string
    }{
        {
            "first_empty_args",
            args{},
            []string{},
        },
        {
            "second_empty_args",
            args{"origin1", map[string]map[string]map[string]map[time.Time]int{}},
            []string{},
        },
        {
            "first_args",
            args{"origin1", map[string]map[string]map[string]map[time.Time]int{"sourceZone1": {"destZone1": {"channel1": {timeArgs1: 6}}}}},
            []string{"insert into ibc_transfer_hourly_stats(zone, zone_src, zone_dest, hour, txs_cnt, period, ibc_channel) values ('origin1', 'sourceZone1', 'destZone1', '2006-01-02T15:04:05', 6, 1, 'channel1')\n    on conflict(zone, zone_src, zone_dest, hour, period, ibc_channel) do update\n        set txs_cnt = ibc_transfer_hourly_stats.txs_cnt + 6;"},
        },
        {
            "second_args",
            args{"origin2", map[string]map[string]map[string]map[time.Time]int{"sourceZone2": {"destZone2": {"channel2": {timeArgs2: 358}}}}},
            []string{"insert into ibc_transfer_hourly_stats(zone, zone_src, zone_dest, hour, txs_cnt, period, ibc_channel) values ('origin2', 'sourceZone2', 'destZone2', '2017-09-11T04:20:49', 358, 1, 'channel2')\n    on conflict(zone, zone_src, zone_dest, hour, period, ibc_channel) do update\n        set txs_cnt = ibc_transfer_hourly_stats.txs_cnt + 358;"},
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addIbcStats(tt.args.origin, tt.args.ibcData)
            assert.Equal(t, tt.expected, actual)
        })
    }
}