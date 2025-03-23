package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type Trade struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Ticker        string    `json:"ticker"`
	Direction     string    `json:"direction"`
	EntryPrice    float64   `json:"entry_price"`
	ExitPrice     float64   `json:"exit_price"`
	Quantity      float64   `json:"quantity"`
	TradeDate     time.Time `json:"trade_date"`
	EntryTime     time.Time `json:"entry_time"`
	ExitTime      time.Time `json:"exit_time"`
	StopLoss      *float64  `json:"stop_loss"`
	TakeProfit    *float64  `json:"take_profit"`
	Commissions   *float64  `json:"commissions"`
	HighestPrice  *float64  `json:"highest_price"`
	LowestPrice   *float64  `json:"lowest_price"`
	Notes         *string   `json:"notes"`
	ScreenshotURL *string   `json:"screenshot_url"`
}

type TradeFilter struct {
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Ticker    string     `json:"ticker"`
	MinProfit *float64   `json:"min_profit"`
	MaxProfit *float64   `json:"max_profit"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	SortBy    string     `json:"sort_by"`
	SortDesc  bool       `json:"sort_desc"`
}

type DbExecutor interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func AddTrade(db DbExecutor, trade Trade) (int, error) {
	stmt, err := db.Prepare(`
        INSERT INTO trades (ticker, direction, entry_price, exit_price, quantity, 
                        trade_date, entry_time, exit_time, stop_loss, take_profit, 
                        commissions, highest_price, lowest_price, notes, screenshot_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) 
        RETURNING id
    `)
	if err != nil {
		log.Printf("Error preparing trade insertion: %v", err)
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(
		trade.Ticker, trade.Direction, trade.EntryPrice, trade.ExitPrice, trade.Quantity,
		trade.TradeDate, trade.EntryTime, trade.ExitTime, trade.StopLoss, trade.TakeProfit,
		trade.Commissions, trade.HighestPrice, trade.LowestPrice, trade.Notes, trade.ScreenshotURL,
	)

	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Printf("Error inserting trade: %v", err)
		return 0, fmt.Errorf("failed to insert trade: %w", err)
	}

	return id, nil
}

func CalculateAndInsertTradeMetrics(db *sql.DB, trade Trade) error {
	if trade.ID == 0 {
		return errors.New("trade ID is required")
	}

	// calculating profit/loss
	var profitLoss float64
	if trade.Direction == "LONG" {
		profitLoss = (trade.ExitPrice - trade.EntryPrice) * trade.Quantity
	} else {
		profitLoss = (trade.EntryPrice - trade.ExitPrice) * trade.Quantity
	}

	// commisions
	if trade.Commissions != nil {
		profitLoss -= *trade.Commissions
	}

	// profit/loss based on percentage of capital risked
	investment := trade.EntryPrice * trade.Quantity
	profitLossPercent := (profitLoss / investment) * 100

	// holding period in mins
	holdingPeriod := int(trade.ExitTime.Sub(trade.EntryTime).Minutes())

	// risk-to-reward and r-multiple
	var riskRewardRatio float64
	var rMultiple float64

	if trade.StopLoss != nil && *trade.StopLoss > 0 {
		// risk per share
		var risk float64
		if trade.Direction == "LONG" {
			risk = trade.EntryPrice - *trade.StopLoss
		} else {
			risk = *trade.StopLoss - trade.EntryPrice
		}

		// calculating actual reward per share
		var reward float64
		if trade.Direction == "LONG" {
			reward = trade.ExitPrice - trade.EntryPrice
		} else {
			reward = trade.EntryPrice - trade.ExitPrice
		}

		if risk > 0 {
			riskRewardRatio = math.Abs(reward / risk)
			rMultiple = reward / risk
		}
	}

	// mfe calculation (how far price moved in your favour)
	var mfe float64
	if trade.Direction == "LONG" && trade.HighestPrice != nil {
		mfe = (*trade.HighestPrice - trade.EntryPrice) * trade.Quantity
	} else if trade.Direction == "SHORT" && trade.LowestPrice != nil {
		mfe = (trade.EntryPrice - *trade.LowestPrice) * trade.Quantity
	}

	// mae calculation (how far price moved against you)
	var mae float64
	if trade.Direction == "LONG" && trade.LowestPrice != nil {
		mae = (trade.EntryPrice - *trade.LowestPrice) * trade.Quantity
	} else if trade.Direction == "SHORT" && trade.HighestPrice != nil {
		mae = (*trade.HighestPrice - trade.EntryPrice) * trade.Quantity
	}

	// Insert metrics into trade_metrics table
	_, err := db.Exec(`
		INSERT INTO trade_metrics 
		(trade_id, profit_loss, profit_loss_percent, risk_reward_ratio, r_multiple, holding_period_minutes, mfe, mae)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (trade_id) 
		DO UPDATE SET 
			profit_loss = EXCLUDED.profit_loss,
			profit_loss_percent = EXCLUDED.profit_loss_percent,
			risk_reward_ratio = EXCLUDED.risk_reward_ratio,
			r_multiple = EXCLUDED.r_multiple,
			holding_period_minutes = EXCLUDED.holding_period_minutes,
			mfe = EXCLUDED.mfe,
			mae = EXCLUDED.mae
	`, trade.ID, profitLoss, profitLossPercent, riskRewardRatio, rMultiple, holdingPeriod, mfe, mae)

	return err
}

func GetTrade(db DbExecutor, id int) (Trade, error) {
	var trade Trade
	stmt, err := db.Prepare(`
		SELECT 
			id, user_id, ticker, direction, entry_price, exit_price, quantity, trade_date, entry_time, 
			exit_time, stop_loss, take_profit, commissions, highest_price, lowest_price, notes, screenshot_url
		FROM trades WHERE id = $1
	`)
	if err != nil {
		return trade, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)
	err = row.Scan(
		&trade.ID, &trade.UserID, &trade.Ticker, &trade.Direction, &trade.EntryPrice, &trade.ExitPrice,
		&trade.Quantity, &trade.TradeDate, &trade.EntryTime, &trade.ExitTime, &trade.StopLoss, &trade.TakeProfit,
		&trade.Commissions, &trade.HighestPrice, &trade.LowestPrice, &trade.Notes, &trade.ScreenshotURL,
	)
	if err == sql.ErrNoRows {
		return trade, fmt.Errorf("trade with ID %d not found", id)
	}
	if err != nil {
		return trade, fmt.Errorf("failed to scan trade: %w", err)
	}

	return trade, nil
}

func ListTrades(db DbExecutor, filter TradeFilter) ([]Trade, error) {
	var conditions []string
	var parameters []interface{}
	parameterIndex := 1

	// apply filters to query
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

	// construct the base query
	query := "SELECT id, user_id, ticker, direction, entry_price, exit_price, quantity, trade_date, entry_time, exit_time, stop_loss, take_profit, commissions, highest_price, lowest_price, notes, screenshot_url FROM trades"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// apply sorting
	if filter.SortBy != "" {
		query += " ORDER BY " + filter.SortBy
		if filter.SortDesc {
			query += " DESC"
		} else {
			query += " ASC"
		}
	}

	// apply limit and offset
	if filter.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(parameterIndex)
		parameters = append(parameters, filter.Limit)
		parameterIndex++
	}
	if filter.Offset > 0 {
		query += " OFFSET $" + strconv.Itoa(parameterIndex)
		parameters = append(parameters, filter.Offset)
	}

	// execute query
	rows, err := db.Query(query, parameters...)
	if err != nil {
		log.Printf("Query execution error: %v", err)
		return nil, fmt.Errorf("failed to retrieve trades: %w", err)
	}
	defer rows.Close()

	// process rows
	var trades []Trade
	for rows.Next() {
		var trade Trade
		err := rows.Scan(
			&trade.ID, &trade.UserID, &trade.Ticker, &trade.Direction, &trade.EntryPrice, &trade.ExitPrice,
			&trade.Quantity, &trade.TradeDate, &trade.EntryTime, &trade.ExitTime, &trade.StopLoss,
			&trade.TakeProfit, &trade.Commissions, &trade.HighestPrice, &trade.LowestPrice,
			&trade.Notes, &trade.ScreenshotURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade row: %w", err)
		}
		trades = append(trades, trade)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return trades, nil
}

func DeleteTrade(db DbExecutor, id int) error {
	stmt, err := db.Prepare("DELETE FROM trades WHERE id = $1")
	if err != nil {
		return fmt.Errorf("failed to delete trade: %w", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete trade: %w", err)
	}
	return nil
}

func UpdateTrade(db DbExecutor, trade Trade) error {
	stmt, err := db.Prepare(`
		UPDATE trades 
		SET 
			ticker = $1, 
			direction = $2, 
			entry_price = $3, 
			exit_price = $4, 
			quantity = $5, 
			trade_date = $6, 
			entry_time = $7, 
			exit_time = $8, 
			stop_loss = $9, 
			take_profit = $10, 
			commissions = $11, 
			highest_price = $12, 
			lowest_price = $13, 
			notes = $14, 
			screenshot_url = $15 
		WHERE id = $16
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		trade.Ticker, trade.Direction, trade.EntryPrice, trade.ExitPrice, trade.Quantity,
		trade.TradeDate, trade.EntryTime, trade.ExitTime, trade.StopLoss, trade.TakeProfit,
		trade.Commissions, trade.HighestPrice, trade.LowestPrice, trade.Notes, trade.ScreenshotURL,
		trade.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update trade: %w", err)
	}

	return nil
}
