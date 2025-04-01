from fastapi import APIRouter, HTTPException, Query, Body
from typing import List

# Import models and services
from app.models.trade_models import (
    Trade, TimePatternResponse, TradeClusterResponse, StrategyEffectivenessResponse
)
from app.services.analysis_service import (
    perform_time_pattern_analysis,
    perform_trade_clustering,
    perform_strategy_effectiveness_analysis
)

# create the api router
router = APIRouter(
    prefix="/analyze",  # use /analyze for the prefix for all python analyzing code
    tags=["Analysis"]   # tag for the openapi documentation
)

@router.post("/time_patterns", response_model=TimePatternResponse)
async def analyze_time_patterns_endpoint(trades: List[Trade] = Body(...)):
    """
    Analyzes trade performance based on the hour of the day and day of the week.
    For trades to be analyzed, they need to have an exit time.
    """
    try:
        result = perform_time_pattern_analysis(trades)
        return result
    except ValueError as ve: # if there are any specific errors, catch them
         raise HTTPException(status_code=400, detail=str(ve))
    except Exception as e:
        print(f"Error in time_patterns endpoint: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during time pattern analysis.")


@router.post("/trade_clusters", response_model=TradeClusterResponse)
async def analyze_trade_clusters_endpoint(
    trades: List[Trade] = Body(...),
    n_clusters: int = Query(3, ge=2, le=20, description="Number of clusters to create"),
    features: List[str] = Query(
        ["duration_seconds", "pnl", "mfe", "mae"],
        description="Features to use for clustering (currently available: duration_seconds, pnl, mfe, mae)"
    )
):
    """
    Groups similar trades using K-Means clustering based on selected features.
    Requires trades to have entry and exit times and prices for feature calculation.
    """
    if not trades:
         raise HTTPException(status_code=400, detail="No trades provided for clustering.")

    try:
        result = perform_trade_clustering(trades, n_clusters, features)
        return result
    except ValueError as ve:
        raise HTTPException(status_code=400, detail=str(ve))
    except Exception as e:
        print(f"Error in trade_clusters endpoint: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during trade clustering.")


@router.post("/strategy_effectiveness", response_model=StrategyEffectivenessResponse)
async def analyze_strategy_effectiveness_endpoint(trades: List[Trade] = Body(...)):
    """
    Analyzes performance metrics for trades grouped by their strategy_tag.
    Trades without a tag are grouped under 'Untagged'.
    """
    if not trades:
        return StrategyEffectivenessResponse(strategy_performance={})

    try:
        result = perform_strategy_effectiveness_analysis(trades)
        return {"strategy_performance": result}
    except Exception as e:
        print(f"Error in strategy_effectiveness endpoint: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during strategy effectiveness analysis.")

