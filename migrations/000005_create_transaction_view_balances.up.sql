CREATE OR REPLACE VIEW transactions_balances AS
SELECT trxb.account_id AS account_id,
       Sum(CASE trxb.type
               WHEN 'credit' THEN trxb.amount
               WHEN 'debit' THEN trxb.amount * -1
           END)        AS balance
FROM (SELECT tr.from_account_id AS account_id,
             'debit'            AS type,
             Sum(tr.amount)     AS amount
      FROM transactions tr
      GROUP BY tr.from_account_id
      UNION ALL
      SELECT tr.to_account_id AS account_id,
             'credit'         AS type,
             Sum(tr.amount)   AS amount
      FROM transactions tr
      GROUP BY tr.to_account_id) AS trxb
GROUP BY trxb.account_id;
