CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(10),
    entry_price DECIMAL(10, 2) CHECK (entry_price > 0),
    exit_price DECIMAL(10, 2) CHECK (exit_price > 0),
    quantity DECIMAL(10, 2), 
    trade_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    stop_loss DECIMAL(10, 2),
    take_profit DECIMAL(10, 2),
    notes TEXT,
    screenshot_url TEXT
);