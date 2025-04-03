from pydantic import BaseModel, Field
from typing import List, Dict, Optional, Any
from datetime import datetime

class SimpleTag(BaseModel): # Renamed from Tag
    name: str
    category: str

class Trade(BaseModel):
    id: int  
    ticker: str
    direction: str
    entry_price: float
    exit_price: Optional[float] = None
    quantity: float
    entry_time: datetime
    exit_time: Optional[datetime] = None
    stop_loss: Optional[float] = None
    take_profit: Optional[float] = None
    commissions: Optional[float] = None
    highest_price: Optional[float] = None # highest price during the trade holding period
    lowest_price: Optional[float] = None  # lowest price during the trade holding period
    notes: Optional[str] = None
    screenshot_url: Optional[str] = None
    pnl: float = Field(..., description="Pre-calculated Profit/Loss for the trade")
    timestamp: Optional[datetime] = Field(None, description="Primary timestamp for analysis (ideally exit_time)")
    strategy_tag: Optional[str] = Field(None, description="Optional tag identifying the strategy used")
    tags: Optional[List[SimpleTag]] = Field(None, description="List of tags (name and category) associated with the trade") # Corrected type hint and description

class TimePerformanceMetrics(BaseModel):
    """The performance metrics we will calculate for each time segment."""
    total_pnl: float = Field(..., description="Sum of PnL for this segment")
    win_rate: float = Field(..., description="Percentage of winning trades (0.0 to 1.0)")
    trade_count: int = Field(..., description="Number of trades in this segment")

class TimePatternResponse(BaseModel):
    """Response for each performance metric in time segment."""
    hourly_performance: Dict[int, TimePerformanceMetrics] = Field(..., description="Performance metrics keyed by hour (0-23)")
    daily_performance: Dict[str, TimePerformanceMetrics] = Field(..., description="Performance metrics keyed by day name (e.g., 'Monday')")

class ClusterInfo(BaseModel):
    """The statistics we will calculate for each cluster when performing K-means."""
    cluster_id: int
    trade_count: int
    avg_pnl: float
    avg_duration_seconds: float
    avg_mfe: float
    avg_mae: float

class TradeClusterResponse(BaseModel):
    """The response for the K-means summary statistics."""
    trade_cluster_map: Dict[int, int] = Field(..., description="Mapping of Trade ID to Cluster ID")
    cluster_summaries: List[ClusterInfo] = Field(..., description="Summary statistics for each cluster")

class StrategyPerformanceMetrics(BaseModel):
    """The performance metrics calculated for each trading strategy."""
    total_pnl: float
    win_rate: float
    profit_factor: Optional[float] = None # can be None if no losses or no profits
    trade_count: int

class StrategyEffectivenessResponse(BaseModel):
    """Response containing performance metrics keyed by strategy tag."""
    strategy_performance: Dict[str, StrategyPerformanceMetrics]


class MarketConditionPerformanceMetrics(BaseModel):
    """Performance metrics for previous day market condition."""
    total_pnl: float
    win_rate: float
    trade_count: int

class MarketCorrelationResponse(BaseModel):
    """Response with calculated performance metrics for each previous day market condition."""
    market_correlation: Dict[str, MarketConditionPerformanceMetrics]
