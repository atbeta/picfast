CREATE TABLE group_strategies (
    group_id      BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    strategy_id   BIGINT NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, strategy_id)
);
