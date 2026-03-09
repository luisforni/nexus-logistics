use axum::{routing::{get, post}, Router, Json, http::StatusCode};
use serde_json::json;

use crate::models::RouteOptimizationRequest;
use crate::optimizer;

pub fn health_routes() -> Router {
    Router::new()
        .route("/health", get(|| async { Json(json!({"status": "ok"})) }))
        .route("/ready",  get(|| async { Json(json!({"status": "ready"})) }))
}

pub fn optimizer_routes() -> Router {
    Router::new()
        .route("/api/v1/optimize/route", post(optimize_route))
}

async fn optimize_route(
    Json(req): Json<RouteOptimizationRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    if req.vehicles.is_empty() {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(json!({"error": "at least one vehicle is required"})),
        ));
    }
    if req.stops.len() > 500 {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(json!({"error": "maximum 500 stops per request"})),
        ));
    }

    let result = tokio::task::spawn_blocking(move || optimizer::solve(&req))
        .await
        .map_err(|e| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(json!({"error": e.to_string()})),
            )
        })?;

    Ok(Json(serde_json::to_value(result).unwrap()))
}
