package models

type TradeStats struct {
	TotalTrades     int     `json:"total_trades"`
	WinningTrades   int     `json:"winning_trades"`
	LosingTrades    int     `json:"losing_trades"`
	TotalProfit     float64 `json:"total_profit"`
	AverageProfit   float64 `json:"average_profit"`
	WinRate         float64 `json:"win_rate"`
	BiggestWin      float64 `json:"biggest_win"`
	BiggestLoss     float64 `json:"biggest_loss"`
	AverageWinSize  float64 `json:"average_win_size"`
	AverageLossSize float64 `json:"average_loss_size"`
	ProfitFactor    float64 `json:"profit_factor"`
}
