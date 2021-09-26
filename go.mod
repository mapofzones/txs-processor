module github.com/mapofzones/txs-processor

go 1.13

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

require (
	github.com/jackc/pgx/v4 v4.6.0
	github.com/mapofzones/cosmos-watcher v0.0.0-20210303220701-2654f0609690
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/go-amino v0.16.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
)
