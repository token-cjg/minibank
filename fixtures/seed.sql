BEGIN;

------------------------------------------------------------------
-- 1.  Companies
------------------------------------------------------------------
INSERT INTO company (company_name)
VALUES  ('Acme Corp'),
        ('Backme Corp')
ON CONFLICT (company_name) DO NOTHING;   -- idempotent re‑runs

------------------------------------------------------------------
-- 2.  Accounts (4 each, random balance 500–10000)
------------------------------------------------------------------
WITH target_companies AS (
    SELECT company_id
    FROM company
    WHERE company_name IN ('Acme Corp', 'Backme Corp')
), gen AS (
    -- generate 4 rows per company_id
    SELECT c.company_id, gs.n
    FROM target_companies c
    CROSS JOIN generate_series(1,4) AS gs(n)
)
INSERT INTO account (company_id, account_balance)
SELECT
    company_id,
    -- random() ∈ [0,1)  ⇒ scale to [500, 10000)
    ROUND((random() * 9500 + 500)::NUMERIC, 2)  -- two decimals
FROM gen
ON CONFLICT DO NOTHING;   -- idempotent re‑runs
------------------------------------------------------------------

COMMIT;
