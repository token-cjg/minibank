CREATE SEQUENCE account_number_seq
  AS BIGINT
  START     WITH 1000000000000000     -- 10^15  (first 16‑digit number)
  INCREMENT BY   1
  MINVALUE  1000000000000000
  MAXVALUE  9999999999999999          -- 10^16‑1
  NO CYCLE;                           --   stop if we ever hit the max

CREATE TABLE company (
  company_id   SERIAL PRIMARY KEY,
  company_name TEXT NOT NULL UNIQUE
);

CREATE TABLE account (
  account_id      BIGSERIAL PRIMARY KEY,
  company_id      INT NOT NULL
                   REFERENCES company(company_id) ON DELETE CASCADE,
  account_number  BIGINT NOT NULL UNIQUE
                   DEFAULT nextval('account_number_seq'),
  account_balance NUMERIC(18,2) NOT NULL DEFAULT 0,
  CHECK (account_balance >= 0),
  CHECK (account_number BETWEEN 1000000000000000 AND 9999999999999999)
);
CREATE INDEX idx_account_number ON account(account_number);  -- fast lookup

CREATE TABLE transaction (
  tx_id              SERIAL PRIMARY KEY,
  source_account_id  INT NOT NULL REFERENCES account(account_id),
  target_account_id  INT NOT NULL REFERENCES account(account_id),
  transfer_amount    NUMERIC(18,2) NOT NULL CHECK (transfer_amount > 0),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_transaction_source ON transaction(source_account_id);
CREATE INDEX idx_transaction_target ON transaction(target_account_id);
