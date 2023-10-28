# banking-system-challenge

## How to run challenge?

Since requirements pointed out `(should not require a pre-installed container/server)` I didn't make a dockerfile, to run the server you need to do the following:

1. Install the required dependencies using the command `go mod download`
2. Start the server using the command `go run cmd/api/main.go`

Default HTTP Port is `8080` if you need to change it change the env var `PORT` by exporting it. ex: `export PORT=9001`

## How to Run Tests for the Go Project

I added a unit testing to validate concurrent safety, as not to have negative balances

1. Run the command `go test ./...`.
2. The test results will be displayed in the command line.

## API Documentation

I added Postman collection you can run commands using it. this project automatically loads accounts from URL was sent in the challenge, and loads it in memory.

Postman Collection: [Open Here](https://api.postman.com/collections/21649836-2405c68d-a986-40ce-9aa6-8183f4fb05e5?access_key=PMAT-01HDHPZ5FAY87DRRC6V04SS81A)

if you don't want to install postman you can run it via curl, these are the Endpoints:

### Get Accounts

you can get all accounts through `[GET] localhost:8080/accounts/`

response:

```json
{
    "accounts": [
        {
            "id": "2ad31c8b-4ee2-4198-85a1-dfb14248fb51",
            "name": "Babbleblab",
            "balance": "4488.1"
        },
        {
            "id": "3a7389f3-d492-4521-ae73-865cb22f7f8a",
            "name": "Skaboo",
            "balance": "365.09"
        },
	....
]}
```

By default this is a highly-available endpoint, it won't gurantee balance consistency in case of an account locked (because of some transaction).

To enable safety you can send `safe` query param as `?safe=true` to have a consistent balance sheets. I added this query param as not to make this endpoint very slow due to waiting to get an account; because it fetches all accounts. (we should add pagination to fix this bottlenick)

curl example to run endpoint:

```
curl --location 'localhost:8080/accounts/?safe=true'
```

### Get Account By ID

you can get a specific account through `[GET] localhost:8080/accounts/:id`

response:

```json
{
    "account": {
        "id": "0a637cbd-5aec-4c3b-8bf0-d8a5eb95024c",
        "name": "Yambee",
        "balance": "3012.9"
    }
}
```

This endpoint is consistant. It'll always return the correct balance no matter what happens. and it does not support unsafe operations.
curl:

```
curl --location 'localhost:8080/accounts/0a637cbd-5aec-4c3b-8bf0-d8a5eb95024c'
```

### Transfer Money

you can transfer money through `[POST] localhost:8080/accounts/:from/transfer/:to`

it returns the sender balance after transfering the money:

```
{
    "Balance": 3002.9
}
```

this endpoint automatically aquires a lock on sender, and reciever accounts before operating, I designed it to return error in case of a failure in transfering, could've added retrying to it with an exponential backoff on every trial to ehnance user experience.

curl:

```
curl --location 'localhost:8080/accounts/0a637cbd-5aec-4c3b-8bf0-d8a5eb95024c/transfer/662178e0-e898-4fa0-a5ac-70951a564f7c' \
--header 'Content-Type: application/json' \
--data '{
    "amount": 10
}'
```

## Scaling & architecture decisions

Currently this service stores data on it's memory, it won't scale this way because it's stateful. I added on `DefaultContext` an interface named `Database` to allow extendable architecture.

```go
type Database[T memorydb.IdentifiedRecord] interface {
	Set(key string, record T, opts memorydb.Opts)
	Setnx(key string, record T) error
	GetM(terms []string, opts memorydb.Opts) []T
	Get(key string, opts memorydb.Opts) (T, error)
	Keys() []memorydb.Key
	Length() int
	Lock(key string) error
	Unlock(key string) error
}
```

In case of we needed to scale out another pod or a replicated node of this service, we can add a package that implements thses methods and talks to any other database over network ex: `Redis`, `Memcached`, `Mongodb`, etc.
