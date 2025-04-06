import pandas as pd
import numpy as np
from typing import List, Dict, Tuple, Optional
from app.models.trade_models import Trade, Tag
import yfinance as yf
import pandas_ta as ta
import datetime
import logging
from collections import defaultdict

# decided to use logging instead of prints for more cleanliness
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

YF_TICKER_MAP = {
    'ES': 'ES=F',
    'GC': 'GC=F',
}
DEFAULT_ATR_PERIOD = 14
DEFAULT_EMA_PERIOD = 21
INDICATOR_LOOKBACK_BUFFER = 60

def fetch_market_data(tickers: List[str], start_date: pd.Timestamp, end_date: pd.Timestamp) -> Dict[str, pd.DataFrame]:
    """
    Determines the min/max range needed for the necessary trades.
    Then, downloads all of the data for all unique tickers.
    """
    market_data = {}
    yf_tickers_map = { YF_TICKER_MAP.get(t, t): t for t in tickers }
    yf_tickers_list = list(yf_tickers_map.keys())

    # add a log to see which ticker and timeframe the data is being downloaded
    logging.info(f"Fetching yfinance data for {len(yf_tickers_list)} tickers from {start_date.date()} to {end_date.date()}")

    try:
        yf_end_date = end_date + pd.Timedelta(days=1)
        data = yf.download(
            yf_tickers_list,
            start=start_date,
            end=yf_end_date,
            progress=False,
            auto_adjust=True,
            group_by='ticker'
        )

        if data.empty:
             logging.warning("yfinance download returned no data.")
             return {ticker: pd.DataFrame() for ticker in tickers}

        # process multi-ticker download result:
        for yf_ticker in yf_tickers_list:
            original_ticker = yf_tickers_map[yf_ticker]
            if yf_ticker in data.columns.levels[0]:
                ticker_data = data[yf_ticker].copy()
                # yfinance might return data with timezone
                if isinstance(ticker_data.index, pd.DatetimeIndex):
                    ticker_data.index = ticker_data.index.tz_localize(None)
                # remove any rows that are missing an open, high, low, or close value
                ticker_data.dropna(subset=['Open', 'High', 'Low', 'Close'], inplace=True)
                if ticker_data.empty:
                    logging.warning(f"No valid data found for {original_ticker} ({yf_ticker}) in the downloaded range.")
                    market_data[original_ticker] = pd.DataFrame()
                else:
                    market_data[original_ticker] = ticker_data
            else:
                logging.warning(f"No data downloaded for ticker {original_ticker} ({yf_ticker}).")
                market_data[original_ticker] = pd.DataFrame()

    except Exception as e:
        logging.error(f"yfinance download failed: {e}")
        return {ticker: pd.DataFrame() for ticker in tickers}

    return market_data


def calculate_indicators(market_data: Dict[str, pd.DataFrame], atr_period: int, ema_period: int) -> Dict[str, pd.DataFrame]:
    """
    Calculates indicators (ATR, EMA for now) for the fetched market data.
    """
    calculated_data = {}
    for ticker, df in market_data.items():
        # make sure all columns have a high, low, and close value
        if df.empty or not all(col in df.columns for col in ['High', 'Low', 'Close']):
            logging.warning(f"Skipping indicator calculation for {ticker} due to missing columns or empty data.")
            calculated_data[ticker] = df
            continue

        try:
            # use pandas technical analysis library to calculate ATR and EMA
            df.ta.atr(length=atr_period, append=True, col_names=('ATR'))
            df.ta.ema(length=ema_period, append=True, col_names=('EMA'))

            # make sure the columns were actually added
            if 'ATR' not in df.columns:
                logging.warning(f"ATR calculation failed for {ticker} (period: {atr_period}). Check data length.")
                df['ATR'] = np.nan
            if 'EMA' not in df.columns:
                logging.warning(f"EMA calculation failed for {ticker} (period: {ema_period}). Check data length.")
                df['EMA'] = np.nan

            calculated_data[ticker] = df
        except Exception as e:
            logging.error(f"Error calculating indicators for {ticker}: {e}")
            df['ATR'] = np.nan
            df['EMA'] = np.nan
            calculated_data[ticker] = df

    return calculated_data


def add_indicator_to_trades(trades_df: pd.DataFrame, market_data_with_indicators: Dict[str, pd.DataFrame]) -> pd.DataFrame:
    """
    Adds the daily indicator values onto the trades DataFrame.
    """
    trades_df = trades_df.copy()
    trades_df['Daily_ATR'] = np.nan
    trades_df['Daily_EMA'] = np.nan
    # extract the date out of the entry time to match the indicator dates
    trades_df['trade_date'] = pd.to_datetime(trades_df['entry_time']).dt.normalize()

    # for each unique ticker in the trades dataframe, get the corresponding table of daily indicators
    for ticker in trades_df['ticker'].unique():
        if ticker not in market_data_with_indicators or market_data_with_indicators[ticker].empty:
            logging.warning(f"No market context data available for ticker {ticker}. Trades for this ticker will have NaN indicators.")
            continue
        ticker_context = market_data_with_indicators[ticker][['ATR', 'EMA']].copy()
        # ake sure the context index is also just the date part
        ticker_context.index = pd.to_datetime(ticker_context.index).normalize()
        # rename columns for clarity
        ticker_context.rename(columns={'ATR': 'Daily_ATR', 'EMA': 'Daily_EMA'}, inplace=True)
        # select trades for the current ticker
        ticker_mask = trades_df['ticker'] == ticker
        # sort both dataframes by date
        trades_subset = trades_df[ticker_mask].sort_values('trade_date')
        ticker_context = ticker_context.sort_index()
        # use merge_asof to merge trade and indicators for that date
        # left_on is the trade date, right_index is the market data date
        merged_data = pd.merge_asof(
            trades_subset,
            ticker_context,
            left_on='trade_date',
            right_index=True,
            direction='backward' # important in case market data might be missing on that day
        )

        # update the original dataframe, and make sure index is aligned
        trades_df.loc[merged_data.index, ['Daily_ATR', 'Daily_EMA']] = merged_data[['Daily_ATR', 'Daily_EMA']].values


    # calculate ema ratio after merging ema values
    trades_df['Daily_EMA21_Ratio'] = np.where(
         pd.notna(trades_df['Daily_EMA']) & (trades_df['Daily_EMA'] != 0),
         (trades_df['entry_price'] / trades_df['Daily_EMA']) - 1,
         np.nan
     )

    return trades_df


def _engineer_basic_features(df: pd.DataFrame) -> pd.DataFrame:
    """Calculates basic trade metrics."""
    df = df.copy()
    # calculate w/l
    df['is_win'] = (df['pnl'] > 0).astype(int)
    # make sure the entry and exit times are datetime objects
    df['entry_time'] = pd.to_datetime(df['entry_time'])
    if 'exit_time' in df.columns:
        df['exit_time'] = pd.to_datetime(df['exit_time'])
        # calculate the duration if both are valid 
        valid_duration = df['exit_time'].notna() & df['entry_time'].notna()
        df['duration_seconds'] = np.nan # initialize the duration column
        df.loc[valid_duration, 'duration_seconds'] = (df.loc[valid_duration, 'exit_time'] - df.loc[valid_duration, 'entry_time']).dt.total_seconds()
    else:
         df['duration_seconds'] = np.nan # just initialize the duration column with no value if the exit time doesn't exist

    df['is_long'] = (df['direction'].str.lower() == 'long').astype(int)
    entry_price = df['entry_price']
    exit_price = df['exit_price']
    stop_loss = df['stop_loss']
    take_profit = df['take_profit']
    quantity = df['quantity']
    direction_mult = np.where(df['is_long'], 1, -1) # 1 for long, -1 for short
    df['profit_margin'] = (direction_mult * (exit_price - entry_price)) / np.where(entry_price != 0, entry_price, np.nan)
    # risk to reward ratio (potential gain / potential loss)
    potential_gain = np.abs(take_profit - entry_price)
    potential_loss = np.abs(entry_price - stop_loss)
    # handle the case where the denominator (potential loss) is 0
    df['risk_reward_ratio'] = np.where(
        potential_loss != 0,
        potential_gain / potential_loss,
        np.nan
    )
    # price excursion (the most a trade went or against your direction)
    # this is assuming highest_price is the max price during trade, and lowest_price is the min price during trade
    df['highest_price_excursion_pct'] = (direction_mult * (df['highest_price'] - entry_price)) / np.where(entry_price != 0, entry_price, np.nan)
    df['lowest_price_excursion_pct'] = (direction_mult * (df['lowest_price'] - entry_price)) / np.where(entry_price != 0, entry_price, np.nan)

    df['normalized_quantity'] = quantity / np.where(entry_price != 0, entry_price, np.nan)

    return df

def _engineer_time_features(df: pd.DataFrame) -> pd.DataFrame:
    """Calculates time-based features."""
    df = df.copy()
    if 'entry_time' not in df.columns:
        logging.warning("Missing 'entry_time' for time feature engineering.")
        return df

    entry_dt = df['entry_time'].dt
    df['day_of_week'] = entry_dt.dayofweek # monday=0, sunday=6
    df['hour_of_day'] = entry_dt.hour
    df['minute_of_hour'] = entry_dt.minute
    # time of day as decimal
    time_decimal = entry_dt.hour + entry_dt.minute / 60.0
    # categorize the time of day as sessions (didn't add london as I only trade new york and asia)
    df['session'] = 'overnight' # default as overnight
    df.loc[(time_decimal >= 20) | (time_decimal < 4), 'session'] = 'asia'
    df.loc[(time_decimal >= 4) & (time_decimal < 9.5), 'session'] = 'premarket'
    df.loc[(time_decimal >= 9.5) & (time_decimal < 16), 'session'] = 'regular'
    df.loc[(time_decimal >= 16) & (time_decimal <= 20), 'session'] = 'afterhours'
    # map the time in a cyclical matter
    seconds_in_day = 24 * 60 * 60
    time_in_seconds = entry_dt.hour * 3600 + entry_dt.minute * 60 + entry_dt.second
    df['time_sin'] = np.sin(2 * np.pi * time_in_seconds / seconds_in_day)
    df['time_cos'] = np.cos(2 * np.pi * time_in_seconds / seconds_in_day)

    # map the day in a cyclical matter
    days_in_week = 7
    df['day_of_week_sin'] = np.sin(2 * np.pi * df['day_of_week'] / days_in_week)
    df['day_of_week_cos'] = np.cos(2 * np.pi * df['day_of_week'] / days_in_week)

    return df

def _engineer_market_context_features(df: pd.DataFrame) -> pd.DataFrame:
    """Calculates features based on market indicators (ATR, EMA for now)."""
    df = df.copy()

    # check if daily_atr column is present
    if 'Daily_ATR' not in df.columns or not df['Daily_ATR'].notna().any():
        logging.warning("Skipping ATR-based features: 'Daily_ATR' column missing or all NaN.")
        return df

    entry_price = df['entry_price']
    exit_price = df['exit_price']
    stop_loss = df['stop_loss']
    take_profit = df['take_profit']
    atr = df['Daily_ATR']
    is_long = df['is_long'].astype(bool)

    # np.where for safe division
    safe_atr = np.where(pd.notna(atr) & (atr != 0), atr, np.nan)
    # price move relative to atr
    price_move = exit_price - entry_price
    df['price_move_atr'] = np.where(is_long, price_move / safe_atr, -price_move / safe_atr)
    # stop distance relative to atr
    stop_dist = entry_price - stop_loss
    df['stop_distance_atr'] = np.where(is_long, stop_dist / safe_atr, -stop_dist / safe_atr) 
    # target distance in atr
    target_dist = take_profit - entry_price
    df['target_distance_atr'] = np.where(is_long, target_dist / safe_atr, -target_dist / safe_atr)
    # mfe (how far price moved in your favor)
    mfe_pct = df['highest_price_excursion_pct'] # make sure this value is >= 0
    # mae (how far price moved against your favor)
    mae_pct = df['lowest_price_excursion_pct']  # this value should be <= 0
    # calculate how much final profit did you get for each unit of maximum drawdown reached? (profit_margin / abs(mae_pct))
    df['profit_to_mae_ratio'] = df['profit_margin'] / np.where(mae_pct != 0, np.abs(mae_pct), np.nan)
    df['profit_to_mae_ratio'].replace([np.inf, -np.inf], np.nan, inplace=True)
    # calculate how much potential shown vs actual risk taken (mfe_pct / abs(mae_pct))
    df['mfe_to_mae_ratio'] = mfe_pct / np.where(mae_pct != 0, np.abs(mae_pct), np.nan)
    df['mfe_to_mae_ratio'].replace([np.inf, -np.inf], np.nan, inplace=True)

    return df

def _encode_categorical_features(df: pd.DataFrame, tags: Optional[List[Tag]] = None) -> pd.DataFrame:
    """Encodes categorical features."""
    df = df.copy()
    # encode tickers (use pd.get_dummies if we implement more tickers in hte future)
    df['ticker_is_ES'] = (df['ticker'] == 'ES').astype(int)
    # encode sessions, use pd.get_dummies 
    if 'session' in df.columns:
        df = pd.get_dummies(df, columns=['session'], prefix='session', dummy_na=False)
    # encode tags
    if 'tags' in df.columns and df['tags'].apply(lambda x: isinstance(x, list) and len(x) > 0).any():
        df['tags'] = df['tags'].apply(lambda x: x if isinstance(x, list) else [])
        sample_tag = None
        for tag_list in df['tags']:
            if tag_list:
                sample_tag = tag_list[0]
                break
        is_dict = isinstance(sample_tag, dict)

        # extract all unique tag names and categories
        all_names = set()
        all_categories = set()
        for tag_list in df['tags']:
            for tag in tag_list:
                try:
                    if is_dict:
                        if 'name' in tag: all_names.add(tag['name'])
                        if 'category' in tag: all_categories.add(tag['category'])
                    else:
                        if hasattr(tag, 'name'): all_names.add(tag.name)
                        if hasattr(tag, 'category'): all_categories.add(tag.category)
                except Exception as e:
                     logging.warning(f"Could not process tag '{tag}': {e}")

        logging.info(f"Found {len(all_names)} unique tag names and {len(all_categories)} unique tag categories.")

        # one-hot encoding for tag names
        for name in sorted(list(all_names)):
            col_name = f'tag_name_{name.replace(" ", "_").lower()}'
            def check_name(tag_list):
                return 1 if any((is_dict and t.get('name') == name) or \
                                (not is_dict and hasattr(t, 'name') and t.name == name)
                                for t in tag_list) else 0
            df[col_name] = df['tags'].apply(check_name)

        # one-hot encoding for categories
        for category in sorted(list(all_categories)):
            col_name = f'tag_cat_{category.replace(" ", "_").lower()}'
            def check_cat(tag_list):
                return 1 if any((is_dict and t.get('category') == category) or \
                                (not is_dict and hasattr(t, 'category') and t.category == category)
                                for t in tag_list) else 0
            df[col_name] = df['tags'].apply(check_cat)
    else:
        logging.info("No 'tags' column found or tags are empty. Skipping tag encoding.")

    return df


def process_trades_for_analysis(
    trades: List[Trade],
    atr_period: int = DEFAULT_ATR_PERIOD,
    ema_period: int = DEFAULT_EMA_PERIOD
) -> pd.DataFrame:
    """
    Main function to process a list of Trade objects into a DataFrame.
    """
    if not trades:
        logging.warning("Input 'trades' list is empty. Returning empty DataFrame.")
        return pd.DataFrame()

    try:
        # 1. convert trades to dataframe
        # 2. prepare for market data fetching
        # 3. fetch market data
        # 4. calculate indicators
        # 5. add indicators for daily context to trades
        # 6. feature engineering pipeline
        # 7. final cleanup 
        df = pd.DataFrame([trade.model_dump() for trade in trades])
        required_cols = ['pnl', 'entry_time', 'ticker', 'entry_price', 'exit_price', 'direction', 'stop_loss', 'take_profit', 'quantity', 'highest_price', 'lowest_price']
        if not all(col in df.columns for col in required_cols):
            missing = [col for col in required_cols if col not in df.columns]
            raise ValueError(f"Input trades missing required fields: {', '.join(missing)}")
        logging.info(f"Initial DataFrame created with {len(df)} trades.")

        # get all of the dates ready for market data fetching
        df['entry_time'] = pd.to_datetime(df['entry_time'])
        min_trade_date = df['entry_time'].min().normalize()
        max_trade_date = df['entry_time'].max().normalize()
        # add buffer for indicator calculation
        fetch_start_date = min_trade_date - pd.Timedelta(days=INDICATOR_LOOKBACK_BUFFER)
        fetch_end_date = max_trade_date
        unique_tickers = df['ticker'].unique().tolist()
        
        market_data_raw = fetch_market_data(unique_tickers, fetch_start_date, fetch_end_date)
        market_data_with_indicators = calculate_indicators(market_data_raw, atr_period, ema_period)
        logging.info("Adding daily market context (indicators) to trades.")
        df = add_indicator_to_trades(df, market_data_with_indicators)
        logging.info("Market context enrichment complete.")
        
        logging.info("Starting feature engineering.")
        df = _engineer_basic_features(df)
        df = _engineer_time_features(df)
        df = _engineer_market_context_features(df)
        df = _encode_categorical_features(df)
        logging.info("Feature engineering complete.")
        
        # drop intermediate columns or columns not needed for analysis
        logging.info(f"Processing finished. Final DataFrame has {df.shape[0]} rows and {df.shape[1]} columns.")
        return df

    except Exception as e:
        logging.error(f"An error occurred during trade processing: {e}", exc_info=True)
        return pd.DataFrame()
