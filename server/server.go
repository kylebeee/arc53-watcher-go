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
	"github.com/kylebeee/arc53-watcher-go/db"
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
	ProviderTypes      []providers.ProviderType
}

const networkMainnet = "mainnet"
const networkTestnet = "testnet"

const indexerMainnetAPI = "https://mainnet-idx.algonode.cloud"
const indexerTestnetAPI = "https://testnet-idx.algonode.cloud"

const algodMainnetAPI = "https://mainnet-api.algonode.cloud"
const algodTestnetAPI = "https://testnet-api.algonode.cloud"

func New() *Arc53WatcherServer {
	var err error

	s := &Arc53WatcherServer{
		Engine:        gin.Default(),
		PrintTxns:     true,
		ProviderTypes: providers.ProviderTypes,
	}

	s.routes()

	network := networkTestnet
	indexerURL := indexerTestnetAPI
	algodURL := algodTestnetAPI
	if s.IsProduction() {
		network = networkMainnet
		indexerURL = indexerMainnetAPI
		algodURL = algodMainnetAPI
	}

	conn, err := db.Connect()
	if err != nil {
		log.Fatalln(err)
	}
	s.DB = conn

	s.Indexer, err = indexer.MakeClient(indexerURL, "")
	if err != nil {
		log.Fatalln(err)
	}

	s.Algod, err = algod.MakeClient(algodURL, "")
	if err != nil {
		log.Fatalln(err)
	}

	var currentAsOfRound int64
	for i := range s.ProviderTypes {
		err = s.ProviderTypes[i].Init(network, s.DB, s.Algod)
		if err != nil {
			log.Fatalf("[!ERR][_MAIN] error initializing provider: %s\n", err)
		}

		startAtRound, err := db.GetLatestProviderRound(s.DB, s.ProviderTypes[i].Type())
		if err != nil {
			log.Fatalf("[!ERR][_MAIN] error fetching provider latest round: %s\n", err)
		}

		err = s.ProviderTypes[i].CatchUp(s.DB, s.Algod, startAtRound, s.Indexer)
		if err != nil {
			log.Fatalf("[!ERR][_MAIN] error catching up provider: %s\n", err)
		}

		if int64(startAtRound) < currentAsOfRound || currentAsOfRound == 0 {
			currentAsOfRound = int64(startAtRound)
		}
	}

	go func() {
		watchingConfig := config.StreamerConfig{
			Algod: &streamer.AlgoConfig{
				FRound: currentAsOfRound,
				LRound: -1,
				Queue:  1,
				ANodes: []*streamer.AlgoNodeConfig{
					{
						Address: algodURL,
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
