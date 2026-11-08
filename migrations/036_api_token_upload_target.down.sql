DROP INDEX IF EXISTS idx_api_tokens_album_id;
DROP INDEX IF EXISTS idx_api_tokens_strategy_id;

ALTER TABLE api_tokens
    DROP COLUMN IF EXISTS album_id,
    DROP COLUMN IF EXISTS strategy_id;
