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
	ctx, cancel := context.WithCancel(context.Background())

	blocks, err := rabbitmq.BlockStream(ctx, os.Getenv("rabbitmq"), "blocks_v2")
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewProcessor(ctx, os.Getenv("postgres"))
	if err != nil {
		log.Fatal(err)
	}

	processor := processor.NewProcessor(ctx, blocks, db)

	err = processor.Process(ctx)

	cancel()
	log.Fatal(err)
}
