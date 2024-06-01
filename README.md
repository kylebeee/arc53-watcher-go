# ARC 53 Watcher

Blockchain watcher for arc53 metadata written in go

## Prerequisites

You'll need go 1.22.1 installed on your local machine. See https://go.dev/ for installation instructions or use Homebrew on MacOS.

- spin up a mysql/mariaDB database
- set an environment variable for the watcher called `DB_AUTH` with your database connection credentials base64 encoded
- create a databases `arc53` and dump the db.sql file into it
```bash
mysql -u username -p arc53 < ./db.sql
```
- fetch all your provider contract IDs and insert them into the database ( you'll do this part yourself by selecting what providers you want to use, currently the `NFD` provider is the only one implemented )
- when its caught up run the watcher to track the chain & update / add new community pages automatically with
```bash
go run ./main/.
```
