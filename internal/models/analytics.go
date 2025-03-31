package models

import (
	"database/sql"
	"errors"
)

type AggregateTradeStats struct {
	TotalTrades          int     `json:"total_trades"`
	WinningTrades        int     `json:"winning_trades"`
	LosingTrades         int     `json:"losing_trades"`
	WinRate              float64 `json:"win_rate"`
	AverageProfitLoss    float64 `json:"average_profit_loss"`
	AverageWinner        float64 `json:"average_winner"`
	AverageLoser         float64 `json:"average_loser"`
	LargestWinner        float64 `json:"largest_winner"`
	LargestLoser         float64 `json:"largest_loser"`
	AverageHoldingPeriod float64 `json:"average_holding_period"`
	ProfitFactor         float64 `json:"profit_factor"`
	ExpectancyPerTrade   float64 `json:"expectancy_per_trade"`
	MaxDrawdown          float64 `json:"max_drawdown"`
	CurrentStreak        int     `json:"current_streak"`
	BreakEvenTrades      int     `json:"break_even_trades"`
}

func GetBasicStats(db *sql.DB, userID int) (AggregateTradeStats, error) {
	var stats AggregateTradeStats
	var tradeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM trades WHERE user_id = $1", userID).Scan(&tradeCount)
	if err != nil {
		return stats, err
	}
	if tradeCount == 0 {
		return stats, errors.New("no trades found for this user")
	}

	// get the total amounts of trades, and win/loss
	err = db.QueryRow(`
		SELECT 
			COUNT(*) as total_trades,
			COUNT(CASE WHEN tm.profit_loss > 0 THEN 1 END) as winning_trades,
			COUNT(CASE WHEN tm.profit_loss < 0 THEN 1 END) as losing_trades,
			COUNT(CASE WHEN tm.profit_loss = 0 THEN 1 END) as break_even_trades
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.TotalTrades, &stats.WinningTrades, &stats.LosingTrades, &stats.BreakEvenTrades)
	if err != nil {
		return stats, err
	}

	// calculating win rate (excluding break-even trades)
	if (stats.TotalTrades - stats.BreakEvenTrades) > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades - stats.BreakEvenTrades)
	}

	// calculate average win/loss and holding period
	err = db.QueryRow(`
		SELECT 
			AVG(tm.profit_loss) as avg_profit_loss,
			AVG(tm.holding_period_minutes) as avg_holding_period
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.AverageProfitLoss, &stats.AverageHoldingPeriod)
	if err != nil {
		return stats, err
	}

	// calculate average winner and loser
	err = db.QueryRow(`
		SELECT 
			COALESCE(AVG(CASE WHEN tm.profit_loss > 0 THEN tm.profit_loss END), 0) as avg_winner,
			COALESCE(AVG(CASE WHEN tm.profit_loss < 0 THEN tm.profit_loss END), 0) as avg_loser
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.AverageWinner, &stats.AverageLoser)
	if err != nil {
		return stats, err
	}

	// find biggest winner and loser
	err = db.QueryRow(`
		SELECT 
			-- use coalesce to handle null values
			COALESCE(MAX(tm.profit_loss), 0) as largest_winner,
			COALESCE(MIN(tm.profit_loss), 0) as largest_loser
		FROM trades t
		-- only for user selected
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.LargestWinner, &stats.LargestLoser)
	if err != nil {
		return stats, err
	}

	// calculate profit factor
	// profit factor is the ratio of total profit to total loss
	err = db.QueryRow(`
		SELECT 
			CASE 
				-- if there are losses, calculate profit factor
				WHEN SUM(CASE WHEN tm.profit_loss < 0 THEN ABS(tm.profit_loss) ELSE 0 END) > 0 
				THEN SUM(CASE WHEN tm.profit_loss > 0 THEN tm.profit_loss ELSE 0 END) / 
					SUM(CASE WHEN tm.profit_loss < 0 THEN ABS(tm.profit_loss) ELSE 0 END)
				-- if no losses but have gains, return a high number to handle division by zero
				WHEN SUM(CASE WHEN tm.profit_loss > 0 THEN tm.profit_loss ELSE 0 END) > 0
				THEN 999.99
				ELSE 0
			END as profit_factor
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.ProfitFactor)
	if err != nil {
		return stats, err
	}

	// calculate expectancy
	stats.ExpectancyPerTrade = (stats.WinRate * stats.AverageWinner) + ((1 - stats.WinRate) * stats.AverageLoser)

	// find if latest trade is a win or loss
	var isLatestTradeWin bool
	err = db.QueryRow(`
		SELECT profit_loss > 0
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
		ORDER BY t.exit_time DESC
		LIMIT 1
	`, userID).Scan(&isLatestTradeWin)

	if err != nil {
		return stats, err
	}

	// find streak length
	err = db.QueryRow(`
		WITH ranked_trades AS (
			SELECT 
				tm.profit_loss > 0 as is_win, -- true if profit_loss is greater than 0 (winning trade)
				ROW_NUMBER() OVER (ORDER BY t.exit_time DESC) as row_num  -- numbers trades from newest to oldest
			FROM trades t
			JOIN trade_metrics tm ON t.id = tm.trade_id
			WHERE t.user_id = $1
			ORDER BY t.exit_time DESC
		),
		-- get the win/loss of the first trade
		first_trade AS (
			SELECT is_win FROM ranked_trades WHERE row_num = 1
		),
		-- count the number of trades in a row with the same win/loss
		SELECT COUNT(*)
		FROM ranked_trades r, first_trade f
		-- check if first trade aligns with the streak
		WHERE r.is_win = f.is_win
		AND r.row_num <= (
			SELECT MIN(r2.row_num) - 1
			FROM ranked_trades r2, first_trade f
			WHERE r2.is_win != f.is_win
			AND r2.row_num > 1
		)
	`, userID).Scan(&stats.CurrentStreak)

	// apply sign based on win/loss
	if !isLatestTradeWin {
		stats.CurrentStreak = -stats.CurrentStreak
	}

	// calculate max drawdown
	// 1. calculate the running balance
	// 2. find the peak balance
	// 3. calculate the drawdown as a percentage of the peak balance
	rows, err := db.Query(`
		-- find running balance by summing up the profit_loss of each trade
		WITH running_balance AS (
			SELECT 
				t.exit_time,
				tm.profit_loss,
				SUM(tm.profit_loss) OVER (ORDER BY t.exit_time) as balance  -- use OVER to calculate the running total at every trade
			FROM trades t
			JOIN trade_metrics tm ON t.id = tm.trade_id
			WHERE t.user_id = $1
			ORDER BY t.exit_time
		),
		-- find the peak balance
		peaks AS (
			SELECT 
				exit_time,
				balance,
				MAX(balance) OVER (ORDER BY exit_time) as peak
			FROM running_balance
		)
		-- calculate the drawdown as a percentage of the peak balance
		SELECT 
			(peak - balance) / NULLIF(peak, 0) * 100 as drawdown_percent
		FROM peaks
		ORDER BY drawdown_percent DESC
		LIMIT 1
	`, userID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&stats.MaxDrawdown)
		if err != nil {
			return stats, err
		}
	}

	return stats, nil
}
