### Getting started

- Install `asdf` and `asdf-golang`
- Run `asdf install golang 1.23.4` to install golang 1.23.4.
- Install postgres `brew install postgres`.
- Start postgres.
- Run `make db_create` && `make db_migrate` to create the database + load the schema.
- Open a db client like DBeaver to view your database `bank`.
- `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`.

Then

- `make run`.

### Running basic checks

- `make lint` to run lint checks
- `make test` to run the test suite
- `make coverage` to run tests + check coverage

### Kicking the tires

- Run `make db_create && make db_migrate` as above
- _Then_ run `make db_seed` to load the example data
- After this, open Postman, and import `minibank.postman_collection.json`.  This should create a new collection in Postman called "Mable".
- Note, unless you are a member of the minibank team, you will need to manually modify the transfer request to pass a new file with the body of the request. Fortunately you are in luck! The fixtures/transfer.csv file provided in this repository should meet all your transference needs:
<img width="1299" alt="transfer_demo" src="https://github.com/user-attachments/assets/1e3b2b1a-9395-4b13-a2c0-b9b90925d6bf" />
- Finally, run `make run` as above to start the server, then go to Postman and experiment with your new requests.
