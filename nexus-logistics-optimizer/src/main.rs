mod optimizer;
mod routes;
mod models;
mod error;

use axum::{Router, middleware};
use tower_http::{
    cors::{Any, CorsLayer},
    trace::TraceLayer,
    compression::CompressionLayer,
    request_id::MakeRequestUuid,
    set_header::SetRequestHeaderLayer,
};
use tracing::info;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

#[tokio::main]
async fn main() -> anyhow::Result<()> {

    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()))
        .with(tracing_subscriber::fmt::layer().json())
        .init();

    let port = std::env::var("OPTIMIZER_PORT").unwrap_or_else(|_| "9090".to_string());
    let addr = format!("0.0.0.0:{port}");

    let app = build_router();

    info!("Nexus Optimizer listening on {addr}");
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;
    Ok(())
}

pub fn build_router() -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .merge(routes::health_routes())
        .merge(routes::optimizer_routes())
        .layer(CompressionLayer::new())
        .layer(TraceLayer::new_for_http())
        .layer(cors)
}
