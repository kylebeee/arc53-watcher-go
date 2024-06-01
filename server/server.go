package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/indexer"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/kylebeee/arc53-watcher-go/errors"
	streamer "github.com/kylebeee/arc53-watcher-go/internal/algod"
	"github.com/kylebeee/arc53-watcher-go/internal/config"
	"github.com/kylebeee/arc53-watcher-go/providers"
)

type Arc53WatcherServer struct {
	*gin.Engine
	Name               errors.Sn
	DB                 *sqlx.DB
	LocalTime          *time.Location
	Algod              *algod.Client
	Indexer            *indexer.Client
	WatcherCancelFn    context.CancelFunc
	ProcessingFailures []interface{}
	PrintTxns          bool
	Providers          []providers.Provider
}

func New() *Arc53WatcherServer {
	var err error

	s := &Arc53WatcherServer{
		Engine:    gin.Default(),
		PrintTxns: true,
		Providers: providers.Providers,
	}

	s.routes()

	s.Indexer, err = indexer.MakeClient("https://mainnet-idx.algonode.cloud", "")
	if err != nil {
		log.Fatalln(err)
	}

	s.Algod, err = algod.MakeClient("https://mainnet-api.algonode.cloud", "")
	if err != nil {
		log.Fatalln(err)
	}

	for i := range s.Providers {
		err = s.Providers[i].Init(s.DB, s.Algod)
		if err != nil {
			log.Fatalf("[!ERR][_MAIN] error initializing provider: %s\n", err)
		}
	}

	go func() {
		watchingConfig := config.StreamerConfig{
			Algod: &streamer.AlgoConfig{
				FRound: -1,
				LRound: -1,
				Queue:  1,
				ANodes: []*streamer.AlgoNodeConfig{
					{
						Address: "https://mainnet-api.algonode.cloud",
						Id:      "public-node",
					},
				},
			},
		}

		var ctx context.Context
		ctx, s.WatcherCancelFn = context.WithCancel(context.Background())

		blocks, status, err := streamer.AlgoStreamer(ctx, watchingConfig.Algod)
		if err != nil {
			log.Fatalf("[!ERR][_MAIN] error getting algod stream: %s\n", err)
		}

		go func() {
			for {
				select {
				case <-status:
					//noop
				case b := <-blocks:
					s.ProcessBlock(b)
				case <-ctx.Done():
					fmt.Println("DONE APPARENTLY")
				}
			}
		}()

		<-ctx.Done()
		fmt.Println("BLOCK WATCHER GOROUTINE FINISHED")
	}()

	return s
}

func (s *Arc53WatcherServer) IsProduction() bool {
	return os.Getenv("ENV") == "production"
}

func (s *Arc53WatcherServer) Close() {
	s.DB.Close()
}
