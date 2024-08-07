# ARC 53 Watcher

Blockchain watcher for arc53 metadata written in go

## Prerequisites

You'll need go 1.22.1 installed on your local machine. See https://go.dev/ for installation instructions or use Homebrew on MacOS.

- spin up a mysql database
- set an environment variable for the watcher called `DB_AUTH` with your database connection credentials base64 encoded
For Linux & MacOS this typically looks like
```bash
 DB_AUTH=`echo -n <database connection string> | base64` && export DB_AUTH
```
> [!NOTE]
> if you're unfamiliar, a connection string typically includes your database username, password, host, port & database ( if necessary ).
>
> ie `<username>:<password>@tcp(<host>:<port>)/`

- set an environment variable for `ENV` if you want to use mainnet
```bash
 export ENV=production
```
- create databases `arc53` & `arc53_test` and dump the db.sql file into them
```bash
mysql -u username -p arc53 < ./db.sql && mysql -u username -p arc53_test < ./db.sql
```

## Running the server
Run the watcher to sync provider apps, track the chain & update / add new community pages automatically using the go run command:
```bash
go run ./main/.
```

> [!NOTE]
> the initial catchup for syncing all provider apps may take some time

## Adding new providers

A provider type in the context of ARC 53 is a type of contract that is capable of doing verifications against multiple addresses & a way to store & retreive the IPFS Content ID which is the location of the JSON metadata contents.

ProviderTypes must implement the following interface:
```golang
type ProviderType interface {
	Type() string
	Init(string, *sqlx.DB, *algod.Client) error
	CatchUp(*sqlx.DB, *algod.Client, uint64, *indexer.Client) error
	ProcessBlock(stxn types.SignedTxnInBlock, round uint64) error
	Process(uint64) error
	IsProviderApp(uint64) bool
}
```

`Type() string` is a namespace label for the provider Type. A provider type is a set of provider contracts that are capable of adhering to the specification where a single provider type interface implementation tracks, saves & updates the details. This allows us to group them by the implementation details inherit in supporting the blockwatcher service for a set of contracts.

`Init(string, *sqlx.DB, *algod.Client) error` is a setup function that implementers can use to instantiate dependencies and ensure the provider type is ready to go.

`CatchUp(*sqlx.DB, *algod.Client, uint64, *indexer.Client) error` is the function that gets ran to do the initial database propegation for the provider type contracts that existed before the server was ran or were created since the last time it was ran.

`ProcessBlock(stxn types.SignedTxnInBlock, round uint64) error` gets ran as the blockwatcher checks for new blocks, in this function provider types must check for new provider apps of their given type & save them, as well as check for updates to existing provider contracts.

`Process(uint64) error` is for one off app updates & allow us to process / update ARC53 data through mechanisms like direct rest api calls

`IsProviderApp(uint64) bool` discerns whether a provided app ID is of a given type
