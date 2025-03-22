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