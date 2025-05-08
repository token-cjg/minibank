### Getting started

- Install `asdf` and `asdf-golang`
- Run `asdf install golang 1.23.4` to install golang 1.23.4.
- Install postgres `brew install postgres`.
- Start postgres.
- Run `make db_create` && `make db_migrate` to create the database + load the schema.
- Open a db client like DBeaver to view your database `bank`.

Then

- `make run`.
