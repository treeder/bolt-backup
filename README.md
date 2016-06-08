
Setting up databases are a pain in the ass. Maintaining databases are a pain in the ass. Especially if you want to 
run a simple microservice or 10. Well how about embed the database and have it backup regularly and auto restore on start? 
The max loss will be the time between the backups. If you can afford to lose a few minutes of data in the rare event of a failure,
this will do the trick. And just think, no database to manage!

This uses [BoltDB](https://github.com/boltdb/bolt), a simple, fast, embeddable database for Go. 

## Usage

First get the lib:

```sh
go get github.com/treeder/bolt-backup
```

Then set your AWS credentials as the following environment variables:

```
AWS_ACCESS_KEY=x
AWS_SECRET_KEY=y
```

Now you can use the Open function, very similar to `bolt.Open`:

```go
package main

import (
    "github.com/treeder/bolt-backup"
)

func main() {
    // The first three params are the same as bolt.Open. 4th is the name of an s3 bucket. 5th is the backup interval.
    db, err := backup.Open("my.db", 0600, nil, "my-backup-bucket", 1*time.Minute)
    // From here, just use bolt as normal.
```

That's all you need to do. 

## Example

See the `/example` directory README for an example you try out, including failure and recovery.
