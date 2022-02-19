package main

import (
	"context"
	"log"
	"os"

	processor "github.com/mapofzones/txs-processor/pkg"
	"github.com/mapofzones/txs-processor/pkg/rabbitmq"
	"github.com/mapofzones/txs-processor/pkg/x/postgres"
)

func main() {
	rabbitmqConnector := os.Getenv("rabbitmq")
	postgresConnector := os.Getenv("postgres")
	queueName := os.Getenv("queue")

	ctx, cancel := context.WithCancel(context.Background())

	blocks, err := rabbitmq.BlockStream(ctx, rabbitmqConnector, queueName)
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewProcessor(ctx, postgresConnector)
	if err != nil {
		log.Fatal(err)
	}

	processor := processor.NewProcessor(ctx, blocks, db)

	err = processor.Process(ctx)

	cancel()
	log.Fatal(err)
}
