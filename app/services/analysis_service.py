import pandas as pd
import numpy as np
from typing import List, Dict, Optional, Any
from datetime import datetime
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer
import yfinance as yf
import pandas_ta as ta
import torch
from torch.utils.data import Dataset, DataLoader
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer

from app.models.trade_models import (
    Tag, Trade, TimePerformanceMetrics, ClusterInfo, StrategyPerformanceMetrics,
    MarketConditionPerformanceMetrics
)

# to make sure we don't have repeated downloads, cache downloaded data
yf_data_cache = {}

def time_pattern_analysis(trades: List[Trade]) -> Dict[str, Dict[Any, TimePerformanceMetrics]]:
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


def trade_clustering(trades: List[Trade], n_clusters: int, features: List[str]) -> Dict[str, Any]:
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

def strategy_effectiveness_analysis(trades: List[Trade]) -> Dict[str, StrategyPerformanceMetrics]:
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


def market_correlation_analysis(trades: List[Trade]) -> Dict[str, MarketConditionPerformanceMetrics]:
    """
    Analyzes trade performance based on previous day's market conditions using yfinance
    (SPY direction and VIX levels for ES, DXY direction and GC=F(utures) direction for XAU/USD(GC)).
    """
    if not trades:
        return {}

    # make sure that ticker and exit time is provided
    valid_trades = [t for t in trades if t.exit_time and t.ticker]
    if not valid_trades:
        print("Warning: No trades with exit_time and ticker provided for market correlation.")
        return {}

    # import all valid trades into a dataframe
    df_trades = pd.DataFrame([trade.model_dump() for trade in valid_trades])
    # make sure that the exit times are treated as datetime values
    df_trades['exit_time'] = pd.to_datetime(df_trades['exit_time'])
    # get only the date from the exit time
    df_trades['exit_date'] = df_trades['exit_time'].dt.floor('D').dt.date 

    # find the earliest and latest dates for all of the trades
    min_date = df_trades['exit_date'].min() - pd.Timedelta(days=7) # include a buffer for shifts and 
    max_date = df_trades['exit_date'].max() + pd.Timedelta(days=1) # include max date

    # fetch market data:
    tickers = ['SPY', '^VIX', 'DX-Y.NYB', 'GC=F']
    try:
        # use yfinance to download the daily open and close for the tickers
        market_data = yf.download(tickers, start=min_date, end=max_date, progress=False) # progress false hides download messages
        if market_data.empty:
            print("Warning: yfinance returned no market data for the specified range.")
            return {}
        # get only the open and close from the market data
        market_data_processed = market_data[['Open', 'Close']].copy()
        # combine the level of columns (Open_SPY, Close_SPY, etc.)
        market_data_processed.columns = ['_'.join(col).strip() for col in market_data_processed.columns.values]
        # in case there are any gaps from holidays or weekends, fill it with the previous day's value using forward fill.
        market_data_processed = market_data_processed.ffill()

    except Exception as e:
        print(f"Error fetching or processing market data from yfinance: {e}")
        return {}

    # calculate the daily market conditions:
    conditions = pd.DataFrame(index=market_data_processed.index) # use the dates from the market data as the index
    # for each ticker, check if they closed higher than it opened
    if 'Close_SPY' in market_data_processed and 'Open_SPY' in market_data_processed:
        conditions['SPY_Up'] = market_data_processed['Close_SPY'] > market_data_processed['Open_SPY']
    if 'Close_^VIX' in market_data_processed:
        conditions['VIX_High'] = market_data_processed['Close_^VIX'] > 20
    if 'Close_DX-Y.NYB' in market_data_processed and 'Open_DX-Y.NYB' in market_data_processed:
        conditions['DXY_Up'] = market_data_processed['Close_DX-Y.NYB'] > market_data_processed['Open_DX-Y.NYB']
    if 'Close_GC=F' in market_data_processed and 'Open_GC=F' in market_data_processed:
        conditions['GCF_Up'] = market_data_processed['Close_GC=F'] > market_data_processed['Open_GC=F']

    # we want to shift the condition data one row down, so that we compare the PREVIOUS day's conditions
    conditions_prev_day = conditions.shift(1)

    # add the exit date as the index for the trades table so we can match each trade by the date
    df_trades['exit_date_dt'] = pd.to_datetime(df_trades['exit_date'])
    df_trades = df_trades.set_index('exit_date_dt')

    # merge the two table by looking up the exit date in the conditions_prev_day dataframe and adding it to each of df_trades' row
    df_merged = pd.merge(df_trades, conditions_prev_day, left_index=True, right_index=True, how='left')

    results = {}
    df_merged['is_win'] = df_merged['pnl'] > 0

    # define which conditions apply to which tickers
    condition_map = {
        'SPY_PrevDay_Up': ('SPY_Up', ['ES']),
        'SPY_PrevDay_Down': ('SPY_Up', ['ES']),
        'VIX_PrevDay_High': ('VIX_High', ['ES']),
        'VIX_PrevDay_Low': ('VIX_High', ['ES']),
        'DXY_PrevDay_Up': ('DXY_Up', ['XAU/USD', 'GC=F', 'XAUUSD']),
        'DXY_PrevDay_Down': ('DXY_Up', ['XAU/USD', 'GC=F', 'XAUUSD']),
        'GCF_PrevDay_Up': ('GCF_Up', ['XAU/USD', 'GC=F', 'XAUUSD']),
        'GCF_PrevDay_Down': ('GCF_Up', ['XAU/USD', 'GC=F', 'XAUUSD']),
    }

    # loop through each condition we want to test
    for result_key, (condition_col, relevant_tickers_patterns) in condition_map.items():
        # skip if something went wrong and condition column doesn't exist
        if condition_col not in df_merged.columns:
            print(f"Warning: Condition column '{condition_col}' not available for analysis.")
            continue

        # see if the condition is positive, and match the relevant tickers
        is_positive_condition = 'Up' in result_key or 'High' in result_key
        ticker_regex = '|'.join(relevant_tickers_patterns)

        # select the condition column from df_merged, and fills NaN values with the opposite of is_positive_condition to make sure
        # that they don't incorrectly match the condition
        # then, compare the column values with is_positive_condition
        condition_mask = df_merged[condition_col].fillna(not is_positive_condition) == is_positive_condition
        # combine the condition mask with a ticker mask to match tickers
        mask = (
            df_merged['ticker'].str.contains(ticker_regex, case=False, na=False, regex=True) &
            condition_mask
        )
        # apply the mask and get the resulting group
        group = df_merged[mask]

        if not group.empty:
            # calculate performance statistics for this group
            win_rate_val = round(group['is_win'].mean(), 2) if not group['is_win'].empty else 0.0
            metrics = MarketConditionPerformanceMetrics(
                total_pnl=round(group['pnl'].sum(), 2),
                win_rate=win_rate_val,
                trade_count=len(group)
            )
            results[result_key] = metrics

    return results


# Tier 4:
# Multi-factor analysis

def _calculate_atr(df, period=14):
    """Helper function to calculate ATR using high, low, and close."""
    # take a dataframe and check if it has high, low, and close prices
    if not all(col in df.columns for col in ['High', 'Low', 'Close']):
        return None
    high_low = df['High'] - df['Low']
    high_close_prev = abs(df['High'] - df['Close'].shift(1))
    low_close_prev = abs(df['Low'] - df['Close'].shift(1))
    # pick the largest value out of hl, hc, and lc
    tr = pd.DataFrame({'hl': high_low, 'hc': high_close_prev, 'lc': low_close_prev}).max(axis=1)
    # find the average true range by taking the average over the specified period
    atr = tr.rolling(window=period, min_periods=period).mean()
    return atr

def add_indicators(trade_row: pd.Series) -> pd.Series:
    """
    Fetches daily market context (ATR, EMA for now) for a single trade using yfinance.
    """
    global yf_data_cache
    ticker_map = {
        'ES': 'ES=F',
        'GC': 'GC=F',
    }
    yf_ticker = ticker_map.get(trade_row['ticker'], trade_row['ticker']) 
    # set the parameters up for downloading enough data for indicators (EMA/ATR)
    trade_date = pd.to_datetime(trade_row['entry_time']).date()
    start_date = trade_date - pd.Timedelta(days=60) 
    end_date = trade_date + pd.Timedelta(days=1)

    # check if the data is already cached. if not, download the data and cache it
    cache_key = (yf_ticker, start_date, end_date)
    if cache_key in yf_data_cache:
        daily_data = yf_data_cache[cache_key]
    else:
        try:
            daily_data = yf.download(yf_ticker, start=start_date, end=end_date, progress=False, auto_adjust=True)
            if daily_data.empty:
                print(f"Warning: No yfinance data for {yf_ticker} from {start_date} to {end_date}")
                return pd.Series({'Daily_ATR': np.nan, 'Daily_EMA21_Ratio': np.nan})
            yf_data_cache[cache_key] = daily_data
        except Exception as e:
            print(f"Warning: yfinance download failed for {yf_ticker}: {e}")
            return pd.Series({'Daily_ATR': np.nan, 'Daily_EMA21_Ratio': np.nan})

    # use pandas_ta to calculate the indicators.
    # I chose to use the 21 EMA because it's what I use when trading ES. however, that's on the 2000 tick chart
    # EMA for general recent trend direction, and ATR for volatility. could add other indicators later, but we'll see
    try:
        daily_data.ta.ema(length=21, append=True) # appends 'EMA_21'
        daily_data.ta.atr(length=14, append=True) # appends 'ATR_14'

        # make sure columns exist before accessing
        ema_col = 'EMA_21'
        atr_col = 'ATR_14'

        if ema_col not in daily_data.columns or atr_col not in daily_data.columns:
             print(f"Warning: Could not calculate EMA/ATR for {yf_ticker}")
             return pd.Series({'Daily_ATR': np.nan, 'Daily_EMA21_Ratio': np.nan})

        # get the indicator values for the trade date
        trade_date_dt = pd.to_datetime(trade_date)
        context_data = daily_data[daily_data.index <= trade_date_dt].iloc[-1:] # get the last row on or before trade date

        if context_data.empty:
            return pd.Series({'Daily_ATR': np.nan, 'Daily_EMA21_Ratio': np.nan})

        atr_value = context_data[atr_col].iloc[0]
        ema21_value = context_data[ema_col].iloc[0]

        # calculate the ema_ratio (the ratio of the trade's entry price, relative to the 21EMA)
        ema_ratio = np.nan
        if pd.notna(ema21_value) and ema21_value != 0:
            ema_ratio = (trade_row['entry_price'] / ema21_value) - 1

        return pd.Series({'Daily_ATR': atr_value, 'Daily_EMA21_Ratio': ema_ratio})

    except Exception as e:
        print(f"Error calculating indicators for {yf_ticker}: {e}")
        return pd.Series({'Daily_ATR': np.nan, 'Daily_EMA21_Ratio': np.nan})


def prepare_data_pattern_recognition(trades: List[Trade], tags: List[Tag]) -> pd.DataFrame:
    """
    Prepares the data for the multi-factor pattern recognition analysis.
    """
    if not trades:
        return pd.DataFrame()

    df = pd.DataFrame([trade.model_dump() for trade in trades])

    if 'pnl' not in df.columns or 'entry_time' not in df.columns or 'ticker' not in df.columns or 'entry_price' not in df.columns:
        raise ValueError("Input trades missing required fields: pnl, entry_time, ticker, entry_price")

    df['is_win'] = (df['pnl'] > 0).astype(int)

    # for each row, add the indicators with the add_indicators function
    # this might take a while, because we potentially have to download and add context row by row
    print("Adding market context to trades through indicators (this may take a while)")
    market_context = df.apply(add_indicators, axis=1)
    df = pd.concat([df, market_context], axis=1)
    print("Enrichment complete.")

    # make sure entry and exit times are in datetime format
    df['entry_time'] = pd.to_datetime(df['entry_time'])
    if 'exit_time' in df.columns:
        df['exit_time'] = pd.to_datetime(df['exit_time'])
        # make sure there is an exit time and entry time to calculate duration
        valid_duration = df['exit_time'].notna() & df['entry_time'].notna()
        df.loc[valid_duration, 'duration_seconds'] = (df.loc[valid_duration, 'exit_time'] - df.loc[valid_duration, 'entry_time']).dt.total_seconds()
        df['duration_seconds'].fillna(np.nan, inplace=True)

    return df


def engineer_features(trades: pd.DataFrame) -> pd.DataFrame:
    """
    Engineers features for the multi-factor pattern recognition analysis.
    """
    # format features
    # numerical features
    trades['duration'] = (trades['exit_time'] - trades['entry_time']).dt.total_seconds()
    trades['profit_margin'] = np.where(trades['direction'] == 'long', (trades['exit_price'] - trades['entry_price']) / trades['entry_price'], (trades['entry_price'] - trades['exit_price']) / trades['entry_price'])
    trades['risk_reward_ratio'] = np.where(trades['direction'] == 'long', (trades['take_profit'] - trades['entry_price']) / (trades['entry_price'] - trades['stop_loss']), (trades['stop_loss'] - trades['entry_price']) / (trades['entry_price'] - trades['take_profit']))
    trades['highest_price_excursion'] = np.where(trades['direction'] == 'long', (trades['highest_price'] - trades['entry_price']) / trades['entry_price'], (trades['entry_price'] - trades['highest_price']) / trades['entry_price'])
    trades['lowest_price_excursion'] = np.where(trades['direction'] == 'long', (trades['lowest_price'] - trades['entry_price']) / trades['entry_price'], (trades['entry_price'] - trades['lowest_price']) / trades['entry_price'])
    trades['normalized_quantity'] = trades['quantity'] / trades['entry_price']

    # categorical features 
    trades['is_long'] = np.where(trades['direction'] == 'long', 1, 0)
    trades['ticker'] = np.where(trades['ticker'] == 'ES', 1, 0)

    # time based features
    trades['day_of_week'] = trades['entry_time'].dt.dayofweek
    trades['hour_of_day'] = trades['entry_time'].dt.hour

    # one hot encoding for tags:
    if 'tags' in trades.columns:
        # fill None values with empty lists to avoid errors
        trades['tags'] = trades['tags'].apply(lambda x: x if isinstance(x, list) else [])

        # extract all unique tag names and categories
        all_names = set()
        all_categories = set()
        for tag_list in trades['tags']:
            if tag_list:
                for tag in tag_list:
                    if 'name' in tag: all_names.add(tag['name'])
                    if 'category' in tag: all_categories.add(tag['category'])
                    elif hasattr(tag, 'name') and hasattr(tag, 'category'):
                         all_names.add(tag.name)
                         all_categories.add(tag.category)


        # create one-hot encoded columns for the tag names
        for name in sorted(list(all_names)): # i just sorted this for consistency
            col_name = f'tag_name_{name.replace(" ", "_").lower()}'
            # create a column for each unique tag across all trades, and apply 1 to the column if it is present in the trade. apply 0 otherwise
            trades[col_name] = trades['tags'].apply(lambda tl: 1 if any( (isinstance(t, dict) and t.get('name') == name) or (hasattr(t, 'name') and t.name == name) for t in tl) else 0)

        # create one-hot encoded columns for the tag categories
        for category in sorted(list(all_categories)):
            col_name = f'tag_cat_{category.replace(" ", "_").lower()}'
            # create a column for each unique category across all trades, and apply 1 for the category if it is present in the trade. apply 0 otherwise
            trades[col_name] = trades['tags'].apply(lambda tl: 1 if any( (isinstance(t, dict) and t.get('category') == category) or (hasattr(t, 'category') and t.category == category) for t in tl) else 0)

    else:
        print("Warning: 'tags' column not found in DataFrame for feature engineering.")


    return trades


def engineer_analysis(trades_df: pd.DataFrame) -> pd.DataFrame:
    
