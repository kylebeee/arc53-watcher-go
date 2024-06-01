package providers

import (
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/providers/nfd"
)

type Provider interface {
	Init(*sqlx.DB, *algod.Client) error
	Process(stxn types.SignedTxnInBlock, round uint64) error
}

var Providers = []Provider{
	&nfd.NFDProvider{},
}
