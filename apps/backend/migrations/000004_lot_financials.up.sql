ALTER TABLE lots
  ADD COLUMN IF NOT EXISTS hammer_price bigint,
  ADD COLUMN IF NOT EXISTS total_due bigint,
  ADD COLUMN IF NOT EXISTS financial_snapshot jsonb,
  ADD COLUMN IF NOT EXISTS winner_user_id uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lots_winner_user_id ON lots(winner_user_id);
