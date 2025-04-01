import pandas as pd
import numpy as np
from typing import List, Dict, Optional
from datetime import datetime
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer

from app.models.trade_models import (
    Trade, TimePerformanceMetrics, ClusterInfo, StrategyPerformanceMetrics
)

def perform_time_pattern_analysis(trades: List[Trade]) -> Dict[str, Dict[Any, TimePerformanceMetrics]]:
    """
    This analyzes trades to see if trading performance depends on the time they happened.
    Specifically, it looks at the hour of the day and the day of the week.
    Trades must have an exit time for this analysis.
    """
    # filter trades to only analyze completed trades for this
    valid_trades = [t for t in trades if t.exit_time]
    if not valid_trades:
        return {"hourly_performance": {}, "daily_performance": {}}

    # use dataframes for easy filtering and nice organization
    # convert valid_trades into a dataframe. model_dump to convert it into dict format, and add exit_time
    df = pd.DataFrame([{**trade.model_dump(), 'timestamp': trade.exit_time} for trade in valid_trades])

    # add new columns to the table based on hour, day, and if the trades won to help grouping.
    df['hour'] = df['timestamp'].dt.hour
    df['day_name'] = df['timestamp'].dt.strftime('%A')
    df['is_win'] = df['pnl'] > 0

    # hourly analysis (group all trades from 6PM in one group, 7PM in another, etc.):
    hourly_grouped = df.groupby('hour')
    # for each hourly group, calculate pnl, trade count, and win rate.
    hourly_analysis = hourly_grouped.agg(
        total_pnl=('pnl', 'sum'),
        trade_count=('pnl', 'count'),
        win_rate=('is_win', 'mean')
    ).round(2)
    # convert results into a dictionary format (keys will be hour groups and values will be aggregate results)
    hourly_performance_raw = hourly_analysis.to_dict(orient='index')
    hourly_performance = {
        int(hour): TimePerformanceMetrics(**metrics)
        for hour, metrics in hourly_performance_raw.items()
    }

    # daily analysis (grouped by monday, tuesday, etc.):
    daily_grouped = df.groupby('day_name')
    # same aggregate calculations as hourly.
    daily_analysis = daily_grouped.agg(
        total_pnl=('pnl', 'sum'),
        trade_count=('pnl', 'count'),
        win_rate=('is_win', 'mean')
    ).round(2)
    # order days so they appear monday - sunday
    days_order = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
    # if no days occurred on a day, it is dropped.
    daily_analysis = daily_analysis.reindex(days_order).dropna(how='all')
    # same as hourly, convert to dict format with days as the keys.
    daily_performance_raw = daily_analysis.to_dict(orient='index') 
    daily_performance = {
        day: TimePerformanceMetrics(**metrics)
        for day, metrics in daily_performance_raw.items()
    }

    return {
        "hourly_performance": hourly_performance,
        "daily_performance": daily_performance
    }


def perform_trade_clustering(trades: List[Trade], n_clusters: int, features: List[str]) -> Dict[str, Any]:
    """
    The goal of the trade clustering analysis is to automatically group similar trades together.
    We can accomplish this using K-means clustering based on specific features.
    """
    required_fields_for_clustering = {
        'entry_time', 'exit_time', 'entry_price', 'exit_price',
        'highest_price', 'lowest_price', 'quantity', 'direction', 'pnl', 'id'
    }

    valid_trades = []
    # for each trade, make sure it has all of the required information
    for t in trades:
        # convert each trade object into a dict
        trade_dict = t.model_dump()
        # make sure fields are not null for the required fields
        if all(trade_dict.get(field) is not None for field in required_fields_for_clustering):
             # make sure that highest >= entry >= lowest for longs, highest >= entry and entry >= lowest for shorts
             is_long = t.direction.lower() == 'long'
             prices_valid = (is_long and t.highest_price >= t.entry_price and t.entry_price >= t.lowest_price) or \
                            (not is_long and t.highest_price >= t.entry_price and t.entry_price >= t.lowest_price)
             if prices_valid:
                 valid_trades.append(t)

    # we need to make sure that we have at least as many trades as the number of clusters
    if len(valid_trades) < n_clusters:
        raise ValueError(f"Not enough valid trades ({len(valid_trades)}) with required fields for clustering into {n_clusters} clusters.")

    # create a dataframe for the valid trades for easier manipulation and organization
    df = pd.DataFrame([trade.model_dump() for trade in valid_trades])

    # calculate new features that might be useful for clustering
    # time the trade was open in seconds
    df['duration_seconds'] = (df['exit_time'] - df['entry_time']).dt.total_seconds()

    # calculate MFE (how much the price moved in your favor during the trade)
    mfe = np.zeros(len(df))
    longs = df['direction'].str.lower() == 'long'
    shorts = ~longs
    # if long, highest - entry. if short, entry - lowest
    mfe[longs] = df.loc[longs, 'highest_price'] - df.loc[longs, 'entry_price']
    mfe[shorts] = df.loc[shorts, 'entry_price'] - df.loc[shorts, 'lowest_price']
    df['mfe'] = mfe.clip(lower=0) # MFE cannot be negative

    # calculate MAE (how much the price moved against you during the trade)
    mae = np.zeros(len(df))
    # if long, entry - lowest. if short, highest - entry
    mae[longs] = df.loc[longs, 'entry_price'] - df.loc[longs, 'lowest_price']
    mae[shorts] = df.loc[shorts, 'highest_price'] - df.loc[shorts, 'entry_price']
    df['mae'] = mae.clip(lower=0) # MAE cannot be negative


    # we'll use the features that the user specifies for the clustering
    available_features = ["duration_seconds", "pnl", "mfe", "mae"]
    # make sure the given featuers are available
    valid_features = [f for f in features if f in available_features]
    if not valid_features:
         raise ValueError("None of the requested features are available.")
    if len(valid_features) != len(features):
         missing = set(features) - set(valid_features)
         print(f"Warning: Requested features not available/calculable and ignored: {missing}")

    df_features = df[valid_features].copy()

    # prepare the data for clustering
    # if there are missing values, we can fill it in with the mean value
    imputer = SimpleImputer(strategy='mean')
    df_features_imputed = imputer.fit_transform(df_features)

    # scale the features so that they have a similar range of values (so that large values don't skew results)
    scaler = StandardScaler()
    df_scaled = scaler.fit_transform(df_features_imputed)

    # clustering
    kmeans = KMeans(n_clusters=n_clusters, random_state=42, n_init=10)
    # train the model on the trades, and assign a data point to the cluster it belongs to
    df['cluster'] = kmeans.fit_predict(df_scaled)

    # create a dictionary to map each tradeID to the assigned cluster
    trade_cluster_map = df.set_index('id')['cluster'].apply(int).to_dict()

    # for each cluster, calculate the summary statistics
    cluster_summaries_list = []
    grouped = df.groupby('cluster')
    for cluster_id, group in grouped:
        summary = ClusterInfo(
            cluster_id=int(cluster_id),
            trade_count=len(group),
            avg_pnl=round(group['pnl'].mean(), 2),
            avg_duration_seconds=round(group['duration_seconds'].mean(), 2),
            avg_mfe=round(group['mfe'].mean(), 2),
            avg_mae=round(group['mae'].mean(), 2)
        )
        cluster_summaries_list.append(summary)

    return {
        "trade_cluster_map": trade_cluster_map,
        "cluster_summaries": cluster_summaries_list
    }


def _calculate_profit_factor(group):
    """Helper to calculate profit factor, also factoring in no loss trades"""
    total_profit = group[group['pnl'] > 0]['pnl'].sum()
    total_loss = abs(group[group['pnl'] < 0]['pnl'].sum())
    if total_loss == 0:
        return float('inf') if total_profit > 0 else None
    return round(total_profit / total_loss, 2)

def perform_strategy_effectiveness_analysis(trades: List[Trade]) -> Dict[str, StrategyPerformanceMetrics]:
    """
    Analyzes performance metrics for trades grouped by their strategy_tag.
    """
    if not trades:
        return {}

    df = pd.DataFrame([trade.model_dump() for trade in trades])

    # fill missing values in the strategy_tag column with "Untagged"
    df['strategy_tag'] = df['strategy_tag'].fillna('Untagged')
    # put true for all winning trades
    df['is_win'] = df['pnl'] > 0

    grouped = df.groupby('strategy_tag')

    strategy_performance = {}
    for name, group in grouped:
        if len(group) == 0: continue
        # calculate performance metrics for each strategy tag group
        metrics = StrategyPerformanceMetrics(
            total_pnl=round(group['pnl'].sum(), 2),
            win_rate=round(group['is_win'].mean(), 2),
            profit_factor=_calculate_profit_factor(group),
            trade_count=len(group)
        )
        strategy_performance[name] = metrics

    return strategy_performance
