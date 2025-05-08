-- 1. Sequence ----------------------------------------------------
CREATE SEQUENCE IF NOT EXISTS account_number_seq
  AS BIGINT
  START WITH 1000000000000000
  INCREMENT BY 1
  MINVALUE 1000000000000000
  MAXVALUE 9999999999999999
  NO CYCLE;

-- 2. Tables ------------------------------------------------------

CREATE TABLE IF NOT EXISTS company (
  company_id   SERIAL PRIMARY KEY,
  company_name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS account (
  account_id      BIGSERIAL PRIMARY KEY,
  company_id      INT NOT NULL
                   REFERENCES company(company_id) ON DELETE CASCADE,
  account_number  BIGINT NOT NULL UNIQUE
                   DEFAULT nextval('account_number_seq'),
  account_balance NUMERIC(18,2) NOT NULL DEFAULT 0,
  CHECK (account_balance >= 0),
  CHECK (account_number BETWEEN 1000000000000000 AND 9999999999999999)
);

CREATE TABLE IF NOT EXISTS transaction (
  tx_id              SERIAL PRIMARY KEY,
  source_account_id  BIGINT REFERENCES account(account_id),
  target_account_id  BIGINT REFERENCES account(account_id),
  transfer_amount    NUMERIC(18,2) NOT NULL CHECK (transfer_amount > 0),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  error              TEXT NULL
);

-- 3. Indexes -----------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_account_number
        ON account(account_number);

CREATE INDEX IF NOT EXISTS idx_transaction_source
        ON transaction(source_account_id);

CREATE INDEX IF NOT EXISTS idx_transaction_target
        ON transaction(target_account_id);
