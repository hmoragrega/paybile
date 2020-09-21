# Payments API Demo project

A simple showcase of a transfer HTTP Rest API written in go.

## Goal
The goal of the exercise is to be able transfer funds between wallets, both between users  

## Interpretation and assumptions
* User can have one or more wallets.
* Transaction are operations that affect the balance of wallets; this allows transfers between users, as the exercise requires, but supports transfer between wallets of the same user too.
* Conversion between currencies has been left out of the scope.
* A simple basic authentication has been used for simplicity, leaving other safer but more complex solutions out (In. ex. expirable tokens and signed requests.)
* This README includes the API documentation, ideally a better doc should be used, for example OpenAPI specs.

## API Usage
### Authentication
The API uses HTTP Basic authentication for simplicity. leaving other safer but more complex solutions out (In. ex. expirable tokens and signed requests.)
```
Authorization: Basic base64("user_login:user_pass")
```
Example for `user A` in the fixtures:
```
Authorization: Basic dXNlcl9hOnVzZXJfYV9wYXNz
```
### List wallet balance
 * Method: `GET`
 * Path: `/api/v1/wallet/{walletID}`

Examples:
```
curl 'http://localhost:8080/api/v1/wallet/2f9b76dd-f689-456e-9080-6789718018a5' \
--header 'Authorization: Basic dXNlcl9hOnVzZXJfYV9wYXNz'
```
```
{
	"id": "2f9b76dd-f689-456e-9080-6789718018a5",
	"user_id": "bbc00191-b064-4655-9075-261ccef978cb",
	"balance": 12.75
}
```
### List wallet transactions
Returns a list of transactions ordered by date.
 * Method: `GET`
 * Path: `/api/v1/wallet/{walletID}/transactions`
 * QueryParameters:
   * `per_page (integer)`: Number of results to return; Default: `20.`
   * `from_id (uuid)`: Option parameters that can be used to select the start transaction for the current page.
   * `order (asc|desc)`: Can be used to select the order of the results. Default: `asc`   

Example:
```
curl 'http://localhost:8080/api/v1/wallet/2f9b76dd-f689-456e-9080-6789718018a5/transactions' \
--header 'Authorization: Basic dXNlcl9hOnVzZXJfYV9wYXNz'
```
```
{
    "results": [
        {
            "id": "9177ad78-e5d5-4d3c-be8c-e0e1f44bbdcc",
            "amount": 20,
            "balance": 20,
            "transaction_type": "deposit",
            "reference_id": null,
            "date": "2020-09-20T10:00:00Z"
        },
        {
            "id": "4bce4401-6b35-4fa1-94b9-ac5ce05d29b1",
            "amount": -7.25,
            "balance": 12.75,
            "transaction_type": "transfer",
            "reference_id": "97ca2b73-7988-4247-82d4-f6ba723a99c9",
            "date": "2020-09-20T11:10:00Z"
        }
    ],
    "next_id": null
}
```
### Transfer funds between wallets
 * Method: `POST`
 * Path: `/api/v1/wallet/{walletID}/transfer`
 * Content-Type: `application/json`
 * RequestBody:
   * `destination_wallet_id (uuid)`: Destination wallet id.
   * `amount (decimal)`: the amount to transfer, cannot be zero or less.
   * `message (string)`: Custom message.

Example:
```
curl -X POST 'http://localhost:8080/api/v1/wallet/2f9b76dd-f689-456e-9080-6789718018a5/transfer' \
--header 'Content-Type: application/json' \
--header 'Authorization: Basic dXNlcl9hOnVzZXJfYV9wYXNz' \
--data-raw '{
    "destination_wallet_id": "4e1d841d-e53f-4785-ba4d-99df05f11eee",
    "amount": 10,
    "message": "Happy Birthday!"
}'
```
```
{
	"id": "d2964584-701a-4525-ba97-4963081a43d4",
	"issuer_id": "bbc00191-b064-4655-9075-261ccef978cb",
	"origin_wallet_id": "2f9b76dd-f689-456e-9080-6789718018a5",
	"destination_wallet_id": "4e1d841d-e53f-4785-ba4d-99df05f11eee",
	"amount": 10,
	"message": "Happy Birthday!",
	"date": "2020-09-21T00:01:58.939549Z"
}
```

## Dev environment
**TL;DR**
```
make up
make fixtures
```

### Boot up
To start the local development environment:
```
make up
```
To rebuild the images
```
make build
```
To stop it:
```
make stop
```
To destroy it:
```
make down
```
### Setup
A set of fixtures can be re-loaded once the DB is ready.
```
make fixtures
```
### Start API service in the host
```
make api
```
### Run tests
Unit tests
```
make unit
```
All tests: unit + integration tests
```
make test
```
