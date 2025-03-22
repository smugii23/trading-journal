package models

import "database/sql"

type TradeStats struct {
	TradeID              int     `json:"trade_id"`
	ProfitLoss           float64 `json:"profit_loss"`
	ProfitLossPercent    float64 `json:"profit_loss_percent"`
	RiskRewardRatio      float64 `json:"risk_reward_ratio"`
	RMultiple            float64 `json:"r_multiple"`
	HoldingPeriodMinutes int     `json:"holding_period_minutes"`
	MFE                  float64 `json:"mfe"`
	MAE                  float64 `json:"mae"`
}

func GetBasicStats(db *sql.DB, userID int) (TradeStats, error) {

}
