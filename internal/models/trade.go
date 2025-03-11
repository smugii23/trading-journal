package models

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
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

type TradeFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Ticker    string
	MinProfit *float64
	MaxProfit *float64
	Limit     int
	Offset    int
	SortBy    string
	SortDesc  bool
}

type DbExecutor interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func AddTrade(db DbExecutor, trade Trade) (int, error) {
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

func GetTrade(db DbExecutor, id int) (Trade, error) {
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
func ListTrades(db DbExecutor, filter TradeFilter) ([]Trade, error) {
	// need to get the condition with the ${parameter index number}
	var conditions []string
	// also need to get the actual parameter value, []interface{} for flexible variable type
	var parameters []interface{}
	parameterIndex := 1
	if filter.StartDate != nil {
		conditions = append(conditions, "trade_date >= $"+strconv.Itoa(parameterIndex))
		parameters = append(parameters, *filter.StartDate)
		parameterIndex++
	}
	if filter.EndDate != nil {
		conditions = append(conditions, "trade_date <= $"+strconv.Itoa(parameterIndex))
		parameters = append(parameters, *filter.EndDate)
		parameterIndex++
	}
	if filter.Ticker != "" {
		conditions = append(conditions, "ticker = $"+strconv.Itoa(parameterIndex))
		parameters = append(parameters, filter.Ticker)
		parameterIndex++
	}
	if filter.MinProfit != nil {
		conditions = append(conditions, "(exit_price - entry_price) * quantity >= $"+strconv.Itoa(parameterIndex))
		parameters = append(parameters, *filter.MinProfit)
		parameterIndex++
	}
	if filter.MaxProfit != nil {
		conditions = append(conditions, "(exit_price - entry_price) * quantity <= $"+strconv.Itoa(parameterIndex))
		parameters = append(parameters, *filter.MaxProfit)
		parameterIndex++
	}
	query := "SELECT id, ticker, entry_price, exit_price, quantity, trade_date, stop_loss, take_profit, notes, screenshot_url FROM trades"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	if filter.SortBy != "" {
		query += " ORDER BY " + filter.SortBy
		if filter.SortDesc {
			query += " DESC"
		} else {
			query += " ASC"
		}
	}
	if filter.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(parameterIndex)
		parameters = append(parameters, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET $" + strconv.Itoa(parameterIndex)
		parameters = append(parameters, filter.Offset)
	}
	rows, err := db.Query(query, parameters...)
	if err != nil {
		log.Printf("Query execution error: %v", err)
		return nil, errors.New("failed to retrieve trades")
	}
	defer rows.Close()
	var trades []Trade
	for rows.Next() {
		var trade Trade
		err := rows.Scan(&trade.ID, &trade.Ticker, &trade.EntryPrice, &trade.ExitPrice,
			&trade.Quantity, &trade.TradeDate, &trade.StopLoss,
			&trade.TakeProfit, &trade.Notes, &trade.Screenshot)
		if err != nil {
			return nil, errors.New("failed to scan trade row")
		}
		trades = append(trades, trade)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("error during row iteration")
	}
	return trades, nil

}

func DeleteTrade(db DbExecutor, id int) error {
	stmt, err := db.Prepare("DELETE FROM trades WHERE id = $1")
	if err != nil {
		return errors.New("failed to delete trade")
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.New("failed to delete trade")
	}
	return nil
}

func UpdateTrade(db DbExecutor, trade Trade) error {
	stmt, err := db.Prepare("UPDATE trades SET ticker = $1, entry_price = $2, exit_price = $3, quantity = $4, trade_date = $5, stop_loss = $6, take_profit = $7, notes = $8, screenshot_url = $9 WHERE id = $10")
	if err != nil {
		return errors.New("failed to prepare update statement")
	}
	defer stmt.Close()
	_, err = stmt.Exec(trade.Ticker, trade.EntryPrice, trade.ExitPrice, trade.Quantity, trade.TradeDate, trade.StopLoss, trade.TakeProfit, trade.Notes, trade.Screenshot, trade.ID)
	if err != nil {
		return errors.New("failed to update trade")
	}
	return nil
}
