import pandas as pd
import numpy as np
from typing import List, Dict, Optional
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, DataLoader
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer
from app.models.trade_models import Trade
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def train_win_predictor(
    trades_df: pd.DataFrame,
    hidden_dims: List[int] = [64, 32],
    lr: float = 0.001,
    epochs: int = 50,
    batch_size: int = 32,
    device: str = 'auto'
) -> Dict:
    """
    Trains a PyTorch model to predict trade success based on features.
    """
    # first, setup the device
    if device == 'auto':
        selected_device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    elif device == 'cuda' and torch.cuda.is_available():
        selected_device = torch.device("cuda")
    else:
        selected_device = torch.device("cpu")
    logging.info(f"Using device: {selected_device}")

    target = 'is_win'
    if target not in trades_df.columns:
        logging.error("Target column '{target}' not found in Dataframe.")
        return {"error": "No suitable feature columns found."}
    if trades_df[target].isnull().any():
        logging.warning("Target column '{target}' contains NaN values. Dropping rows with NaN values.")
        trades_df = trades_df.dropna(subset=[target])
        if trades_df.empty:
            logging.error("Dataframe is emptyy after dropping all NaN target values.")
            return {"error": "No valid data after handling NaN targets."}
    # create a list containing only the names of columns that are numerical, or not explicitly excluded
    feature_cols = [
        col for col in trades_df.columns if col not in [
            target, 'trade_id', 'entry_time', 'exit_time', 'trade_date',
        'pnl', 'profit_margin', 'tags', 'session',
        ] and  trades_df[col].dtype in [
            np.int64, np.float64, np.int32, np.float32, np.uint8, bool
        ] and not trades_df[col].isnull().all()
    ]   
    if not feature_cols:
        logging.error("No suitable feature columns found after filtering.")
        return {"error": "No suitable feature columns found."}
    logging.info(f"Selected {len(feature_cols)} features: {feature_cols}")
    
    # create dataframes for x (input data / features) and y (target variable / labels), in this case w/l
    X = trades_df[feature_cols]
    y = trades_df[target].astype(int)   
    logging.info(f"Data selected: X shape {X.shape}, y shape {y.shape}")

    # imputation to fill blanks, i'm going to use median to counter outliers
    imputer = SimpleImputer(strategy='median')
    X_imputed = imputer.fit_transform(X)
    X_imputed_df = pd.DataFrame(X_imputed, columns = feature_cols, index=X.index)
    logging.info("Imputation complete.")

    # scale features so that they don't give a bias
    scaler = StandardScaler()
    x_scaled = scaler.fit_transform(X_imputed_df)
    logging.info("Scaling complete.")

    