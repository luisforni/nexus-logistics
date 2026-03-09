use thiserror::Error;
use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Bad request: {0}")]
    BadRequest(String),

    #[error("Internal error: {0}")]
    Internal(#[from] anyhow::Error),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound(msg)   => (StatusCode::NOT_FOUND,            msg.clone()),
            AppError::BadRequest(msg) => (StatusCode::BAD_REQUEST,          msg.clone()),
            AppError::Internal(e)     => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()),
        };

        (status, Json(json!({ "error": message }))).into_response()
    }
}
