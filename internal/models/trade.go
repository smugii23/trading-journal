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

type DbExecutor interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func addTrade(db DbExecutor, trade Trade) (int, error) {
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

func getTrade(db DbExecutor, id int) (Trade, error) {
	var trade Trade
	stmt, err := db.Prepare(`SELECT id, ticker, entry_price, exit_price, quantity, trade_date, 
        stop_loss, take_profit, notes, screenshot_url
 		FROM trades WHERE id = $1`)
	if err != nil {
		return trade, errors.New("failed to retrieve trade")
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	err = row.Scan(&trade.ID, &trade.Ticker, &trade.EntryPrice, &trade.ExitPrice,
		&trade.Quantity, &trade.TradeDate, &trade.StopLoss,
		&trade.TakeProfit, &trade.Notes, &trade.Screenshot)
	if err == sql.ErrNoRows {
		return trade, errors.New("trade not found")
	}
	if err != nil {
		return trade, errors.New("failed to retrieve trade")
	}
	return trade, nil
}
