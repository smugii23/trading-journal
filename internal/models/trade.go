package models

import (
	"database/sql"
	"errors"
	"time"
)

type Trade struct {
	ID         int
	Ticker     string
	EntryPrice float64
	ExitPrice  float64
	Quantity   float64
	TradeDate  time.Time
	StopLoss   float64
	TakeProfit float64
	Notes      string
	Screenshot string
}

func addTrade(db *sql.DB, trade Trade) (int, error) {
	stmt, err := db.Prepare("INSERT INTO trades (ticker, entry_price, exit_price, quantity, trade_date, stop_loss, take_profit, notes, screenshot_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id")
	if err != nil {
		return 0, errors.New("failed to insert trade")
	}
	defer stmt.Close()
	row := stmt.QueryRow(trade.Ticker, trade.EntryPrice, trade.ExitPrice, trade.Quantity, trade.TradeDate, trade.StopLoss, trade.TakeProfit, trade.Notes, trade.Screenshot)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return 0, errors.New("failed to insert trade")
	}
	return id, nil
}
