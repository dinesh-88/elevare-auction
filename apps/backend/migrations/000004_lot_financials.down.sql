ALTER TABLE lots
  DROP COLUMN IF EXISTS winner_user_id,
  DROP COLUMN IF EXISTS financial_snapshot,
  DROP COLUMN IF EXISTS total_due,
  DROP COLUMN IF EXISTS hammer_price;
