
"""
nexus-logistics-optimizer/python/demand_forecast.py

Demand forecasting microservice using Facebook Prophet.
Exposes a FastAPI endpoint that the Rust optimizer calls for AI-driven
demand signals.

Usage:
    uvicorn demand_forecast:app --host 0.0.0.0 --port 9091
"""

from __future__ import annotations

import logging
from datetime import datetime

import pandas as pd
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from prophet import Prophet

logger = logging.getLogger("nexus-forecast")
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
)

app = FastAPI(
    title="Nexus Demand Forecast API",
    version="1.0.0",
    description="AI-driven demand forecasting for supply chain optimization",
)

class DataPoint(BaseModel):
    date: str = Field(..., description="ISO 8601 date string, e.g. 2026-03-06")
    shipments: float = Field(..., ge=0, description="Observed shipment volume")

class ForecastRequest(BaseModel):
    history: list[DataPoint] = Field(
        ...,
        min_length=14,
        description="At least 14 historical data points",
    )
    horizon_days: int = Field(
        default=30,
        ge=1,
        le=365,
        description="Forecast horizon in days",
    )
    include_uncertainty: bool = Field(default=True)

class ForecastPoint(BaseModel):
    date: str
    predicted: float
    lower: float
    upper: float

class ForecastResponse(BaseModel):
    horizon_days: int
    generated_at: str
    forecast: list[ForecastPoint]
    trend_direction: str

@app.post(
    "/api/v1/forecast",
    response_model=ForecastResponse,
    summary="Demand forecast",
)
async def forecast(req: ForecastRequest) -> ForecastResponse:
    """
    Fit a Prophet model on historical shipment data and return a multi-step
    demand forecast with uncertainty intervals.
    """
    try:
        df = pd.DataFrame([
            {"ds": pd.to_datetime(p.date), "y": p.shipments}
            for p in req.history
        ]).sort_values("ds")
    except Exception as exc:
        raise HTTPException(
            status_code=422,
            detail=f"Invalid date format: {exc}",
        ) from exc

    if df["y"].isnull().any():
        raise HTTPException(
            status_code=422,
            detail="History contains null shipment values",
        )

    model = Prophet(
        yearly_seasonality=True,
        weekly_seasonality=True,
        daily_seasonality=False,
        interval_width=0.95,
        changepoint_prior_scale=0.05,
    )
    model.fit(df)

    future = model.make_future_dataframe(periods=req.horizon_days, freq="D")
    forecast_df = model.predict(future)

    future_rows = forecast_df[forecast_df["ds"] > df["ds"].max()].copy()
    for col in ("yhat", "yhat_lower", "yhat_upper"):
        future_rows[col] = future_rows[col].clip(lower=0)

    forecast_points = [
        ForecastPoint(
            date=row["ds"].strftime("%Y-%m-%d"),
            predicted=round(row["yhat"], 2),
            lower=round(row["yhat_lower"], 2),
            upper=round(row["yhat_upper"], 2),
        )
        for _, row in future_rows.iterrows()
    ]

    if len(forecast_points) >= 2:
        delta = forecast_points[-1].predicted - forecast_points[0].predicted
        trend = "up" if delta > 5 else "down" if delta < -5 else "flat"
    else:
        trend = "flat"

    return ForecastResponse(
        horizon_days=req.horizon_days,
        generated_at=datetime.utcnow().isoformat() + "Z",
        forecast=forecast_points,
        trend_direction=trend,
    )

@app.get("/health", summary="Health check")
async def health() -> dict[str, str]:
    return {"status": "ok"}
