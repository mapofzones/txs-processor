package postgres

// queries that write to db

const addZoneQuery = `insert into zones(name, chain_id, is_enabled, is_caught_up) values %s
    on conflict (chain_id) do update
        set is_enabled = %t;`

const addImplicitZoneQuery = `insert into zones(name, chain_id, is_enabled, is_caught_up) values %s
    on conflict (chain_id) do nothing;`

const markBlockQuery = `insert into blocks_log(zone, last_processed_block, last_updated_at) values %s
    on conflict (zone) do update
        set last_processed_block = blocks_log.last_processed_block + 1,
            last_updated_at = '%s';`

const addTxStatsQuery = `insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values %s
    on conflict (hour, zone, period) do update
        set txs_cnt = total_tx_hourly_stats.txs_cnt + %d,
			txs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + %d,
			txs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + %d,
            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + %d;`

const addActiveAddressesQuery = `insert into active_addresses(address, zone, hour, period, is_internal_tx, is_internal_transfer, is_external_transfer) values %s
    on conflict (address, zone, hour, period) do update
        set is_internal_tx = active_addresses.is_internal_tx or EXCLUDED.is_internal_tx,
			is_internal_transfer = active_addresses.is_internal_transfer or EXCLUDED.is_internal_transfer,
			is_external_transfer = active_addresses.is_external_transfer or EXCLUDED.is_external_transfer;`

const addClientsQuery = `insert into ibc_clients(zone, client_id, chain_id) values %s
    on conflict (zone, client_id) do nothing;`

const addConnectionsQuery = `insert into ibc_connections(zone, connection_id, client_id) values %s
    on conflict (zone, connection_id) do nothing;`

const addChannelsQuery = `insert into ibc_channels(zone, channel_id, connection_id, is_opened) values %s
    on conflict(zone, channel_id) do nothing;`

const markChannelQuery = `update ibc_channels
    set is_opened = %t
        where zone = '%s'
        and channel_id = '%s';`

const addIbcStatsQuery = `insert into ibc_transfer_hourly_stats(zone, zone_src, zone_dest, hour, txs_cnt, period, ibc_channel, txs_fail_cnt) values %s
    on conflict(zone, zone_src, zone_dest, hour, period, ibc_channel) do update
        set txs_cnt = ibc_transfer_hourly_stats.txs_cnt + %d,
            txs_fail_cnt = ibc_transfer_hourly_stats.txs_fail_cnt + %d;`

const addIbcCashflowQuery = `insert into ibc_transfer_hourly_cashflow(zone, zone_src, zone_dest, hour, period, ibc_channel, denom, amount) values %s
    on conflict(zone, zone_src, zone_dest, hour, period, ibc_channel, denom) do update
        set amount = ibc_transfer_hourly_cashflow.amount + %d;`

// read-only queries

const lastProcessedBlockQuery = `select last_processed_block from blocks_log
    where zone = '%s';`

const chainIDFromClientIDQuery = `select chain_id from ibc_clients
	where client_id = '%s'
		and zone = '%s';`

const clientIDFromConnectionIDQuery = `select client_id from ibc_connections
	where connection_id = '%s'
		and zone = '%s';`

const connectionIDFromChannelIDQuery = `select connection_id from ibc_channels
	where channel_id = '%s'
		and zone = '%s';`

const statusFromChannelIDQuery = `select is_opened from ibc_channels
	where channel_id = '%s'
		and zone = '%s';`
