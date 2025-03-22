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
			COUNT(CASE WHEN tm.profit_loss <= 0 THEN 1 END) as losing_trades
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.TotalTrades, &stats.WinningTrades, &stats.LosingTrades)
	if err != nil {
		return stats, err
	}

	// calculating win rate
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades)
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
			COALESCE(MAX(tm.profit_loss), 0) as largest_winner,
			COALESCE(MIN(tm.profit_loss), 0) as largest_loser
		FROM trades t
		JOIN trade_metrics tm ON t.id = tm.trade_id
		WHERE t.user_id = $1
	`, userID).Scan(&stats.LargestWinner, &stats.LargestLoser)
	if err != nil {
		return stats, err
	}

	// calculate profit factor
	err = db.QueryRow(`
		SELECT 
			CASE 
				WHEN SUM(CASE WHEN tm.profit_loss < 0 THEN ABS(tm.profit_loss) ELSE 0 END) > 0 
				THEN SUM(CASE WHEN tm.profit_loss > 0 THEN tm.profit_loss ELSE 0 END) / 
					 SUM(CASE WHEN tm.profit_loss < 0 THEN ABS(tm.profit_loss) ELSE 0 END)
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

	// find the current streak
	err = db.QueryRow(`
		WITH ranked_trades AS (
			SELECT 
				tm.profit_loss,
				ROW_NUMBER() OVER (ORDER BY t.exit_time DESC) as row_num
			FROM trades t
			JOIN trade_metrics tm ON t.id = tm.trade_id
			WHERE t.user_id = $1
			ORDER BY t.exit_time DESC
		)
		SELECT 
			(SELECT COUNT(*) 
			 FROM ranked_trades 
			 WHERE row_num <= (SELECT MIN(row_num) FROM ranked_trades WHERE profit_loss * (SELECT SIGN(profit_loss) FROM ranked_trades WHERE row_num = 1) <= 0)
			 AND profit_loss * (SELECT SIGN(profit_loss) FROM ranked_trades WHERE row_num = 1) > 0) 
			* SIGN((SELECT profit_loss FROM ranked_trades WHERE row_num = 1)) as current_streak
		FROM ranked_trades
		LIMIT 1
	`, userID).Scan(&stats.CurrentStreak)
	if err != nil {
		return stats, err
	}

	// calculate max drawdown
	rows, err := db.Query(`
		WITH running_balance AS (
			SELECT 
				t.exit_time,
				tm.profit_loss,
				SUM(tm.profit_loss) OVER (ORDER BY t.exit_time) as balance
			FROM trades t
			JOIN trade_metrics tm ON t.id = tm.trade_id
			WHERE t.user_id = $1
			ORDER BY t.exit_time
		),
		peaks AS (
			SELECT 
				exit_time,
				balance,
				MAX(balance) OVER (ORDER BY exit_time) as peak
			FROM running_balance
		)
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
