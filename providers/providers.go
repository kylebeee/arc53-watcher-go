package providers

import (
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/providers/nfd"
)

type ProviderType interface {
	Type() string
	Init(string, *sqlx.DB, *algod.Client) error
	CatchUp(*sqlx.DB, *algod.Client, uint64, *indexer.Client) error
	ProcessBlock(stxn types.SignedTxnInBlock, round uint64) error
	Process(uint64) error
}

var ProviderTypes = []ProviderType{
	&nfd.NFDProvider{},
}
