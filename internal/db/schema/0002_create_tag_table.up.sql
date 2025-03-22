CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    color VARCHAR(20),
    UNIQUE(user_id, name)
);

CREATE TABLE trade_tags (
    trade_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (trade_id, tag_id),
    FOREIGN KEY (trade_id) REFERENCES trades(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

ALTER TABLE trades
ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1,
ADD COLUMN direction VARCHAR(5) NOT NULL DEFAULT 'LONG' CHECK (direction IN ('LONG', 'SHORT')),
ADD COLUMN entry_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN exit_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN commissions DECIMAL(10, 2) DEFAULT 0,
ADD COLUMN highest_price DECIMAL(10, 2),
ADD COLUMN lowest_price DECIMAL(10, 2);

CREATE TABLE trade_metrics (
    trade_id INTEGER PRIMARY KEY,
    profit_loss DECIMAL(10, 2),
    profit_loss_percent DECIMAL(6, 2),
    risk_reward_ratio DECIMAL(6, 2),
    r_multiple DECIMAL(5, 2),
    holding_period_minutes INTEGER,
    mfe DECIMAL(10, 2),
    mae DECIMAL(10, 2),
    FOREIGN KEY (trade_id) REFERENCES trades(id) ON DELETE CASCADE
);