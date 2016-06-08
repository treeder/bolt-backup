

## Build and run

```sh
go build
AWS_ACCESS_KEY=X AWS_SECRET_KEY=Y AWS_BUCKET_NAME=bolt-backups ./example
```

Then post some data:

```sh
curl -d '{"name":"johnny2"}' http://localhost:8080/users
```

Then check data:

```sh
curl http://localhost:8080/users
```

## Test failure/recover

Go ahead and stop the service (ctrl-c) and delete the database file, `my.db`. 

Now run the startup command above again, then check data with the curl command above.

Boom. Magic.
