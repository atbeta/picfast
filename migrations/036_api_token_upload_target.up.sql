ALTER TABLE api_tokens
    ADD COLUMN strategy_id BIGINT REFERENCES strategies(id) ON DELETE SET NULL,
    ADD COLUMN album_id BIGINT REFERENCES albums(id) ON DELETE SET NULL;

CREATE INDEX idx_api_tokens_strategy_id ON api_tokens(strategy_id);
CREATE INDEX idx_api_tokens_album_id ON api_tokens(album_id);
