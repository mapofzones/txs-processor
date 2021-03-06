package processor

import (
	"context"

	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
)

type Processor interface {
	Handler
	// Validate is used to make all necessary checks before processing block,
	// such as that block received is valid and block processing order
	// is not messed up
	Validate(context.Context, watcher.Block) error
	// commit is used to transact all state changes if that is necessary
	Commit(context.Context, watcher.Block) error
}
