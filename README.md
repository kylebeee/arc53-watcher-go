# arc53-watcher-go

Blockchain watcher for arc53 metadata written in go

- spin up a mysql/mariaDB database
- set an environment variable for the watcher called `DB_AUTH` with your database connection credentials base64 encoded
- run the db.sql file into it to create your tables
- fetch all your provider contract IDs and insert them into the database ( you'll do this part yourself depending on what providers you want to support )
- when its caught up run the watcher to track the chain & update / add new community pages automatically
