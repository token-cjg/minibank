## Minibank

An account + transaction management system for all of your incoming + outgoing payment related needs.

## Business requirements

See [here](./problem_specification/MABLE_BACK_END_CODE_TEST.md) for the business requirements.

## Approach

I opted to use gorilla mux to write a lightweight server, backed by Postgres.

### Entity modelling / Data structure

For entity modelling I opted for three main entities:

- Companies. These are simple things, basically just a `company_name` and an id.
- Accounts. A bit more nuanced. Trap for the young player - one needs to have a 16 digit number for _each and every one of these things_. To action this I decided simply to autoincrement starting from the smallest 16 digit number, with a no cycle condition on the largest 16 digit number. Note too that these are scoped to a given company. i.e. there is a 1 to N relationship between Companies and Accounts.  So I decided to model the data here as an `id`, a `company_id`, an `account_number`, and an `account_balance`.
- Transactions. This is a record of transactions between accounts. I thought it would be useful to retain this information, as it would be useful for sense-checking + also re-running transfers if there was some form of server error or other issue. Attributes decided upon:

```
  tx_id              SERIAL PRIMARY KEY,
  source_account_id  BIGINT NOT NULL REFERENCES account(account_id),
  target_account_id  BIGINT NOT NULL REFERENCES account(account_id),
  transfer_amount    NUMERIC(18,2) NOT NULL CHECK (transfer_amount > 0),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  error              TEXT NULL
```

Note that there is a constraint on the `transfer_amount`, it must be positive.

### Application structure + business logic overview

After these basic decisions, I actioned the following:

- MVC approach for API requests
  1) Models (data structures / entities)
  2) Controllers/Handlers (API endpoints)
  3) Repository/Server layer (Business logic)
- Basic APIs to Create, List, and Get accounts and companies,
- API to batch process txns per inputted csv data

I opted to allow for both `text/csv` and `multipart` form data for csvs. Note I only tested the latter, and debated with myself whether to remove the former but opted to retain it. `multipart` data is processed in a batch, wherein the csv is scanned row by row, and requests are run against the database to compute whether a transfer can occur.

If the balance were to drop below zero I do not action the transfer, but still record the txn, with an entry in the `error` column indicating insufficient funds. A couple of examples can be viewed in the screenshot in [Dev.md](Dev.md).

### Best practices followed

Naturally a few other principles are followed here:

- Unit tests. Only a touch over 60% overall coverage but I think okay as a starting point.
- Linting. I disabled the `errcheck` linting option, but this should be moderately straightforward to fix. I've created a task against the repo to action this.
- Doc generation. I've introduced some basic tooling to create + serve docs for this app.
- CI. Tests and Linting are run in github actions against each and every candidate pull request.

### Potential enhancements

- I've endeavoured to ensure that there is not too much complexity at any level of the application. There are ways of course to safeguard against this automatically, see the issue I've created to tackle this in CI.
- Wherever possible I have sought to follow restful conventions for APIs, however `transfer` is a bit of an exception. I've created a task to explore improving on this slightly.
- I've sought to put in place basic error handling wherever possible. Naturally this would be enhanced by fixing the `errcheck` linting issues as this would point to spots in the app that could be further improved.
- Batching transactions is moderately efficient, but for truly large input files I could look either into leveraging goroutines, or alternatively look into introducing the idea of a `Job` entity and have some microservice that consumes these jobs + actions them, i.e. a `txn_worker`. This would be useful if the app needed to scale to handling large data files with a relatively high frequency, rather than a sporadic batch process.

## Enough lollygagging, just show me how to run the app already

See [here](Dev.md).
