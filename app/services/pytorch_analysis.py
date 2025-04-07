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

def train_pytorch_win_predictor(
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
        logging.error("Target column '{target}' not found in DataFrame.")
        return {"error": "No suitable feature columns found."}
    if trades_df[target].isnull().any():
        logging.warning("Target column '{target}' contains NaN values. Dropping rows with NaN values.")
        trades_df = trades_df.dropna(subset=[target])
        if trades_df.empty:
            logging.error("Dataframe is emptyy after dropping all NaN target values.")
            return {"error": "No valid data after handling NaN targets."}
    
    