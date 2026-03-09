use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Coordinate {
    pub lat: f64,
    pub lng: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Stop {
    pub id: Uuid,
    pub label: String,
    pub coordinate: Coordinate,

    pub time_window: Option<(u32, u32)>,

    pub service_minutes: u32,

    pub demand: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Vehicle {
    pub id: Uuid,
    pub capacity: f64,
    pub max_stops: usize,
}

#[derive(Debug, Deserialize)]
pub struct RouteOptimizationRequest {
    pub depot: Coordinate,
    pub stops: Vec<Stop>,
    pub vehicles: Vec<Vehicle>,
    pub objective: ObjectiveFunction,
}

#[derive(Debug, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub enum ObjectiveFunction {
    #[default]
    MinimizeDistance,
    MinimizeTime,
    BalanceLoad,
}

#[derive(Debug, Serialize)]
pub struct Route {
    pub vehicle_id: Uuid,
    pub stop_sequence: Vec<Uuid>,
    pub total_distance_km: f64,
    pub total_minutes: f64,
    pub total_demand: f64,
}

#[derive(Debug, Serialize)]
pub struct OptimizationResult {
    pub routes: Vec<Route>,
    pub total_distance_km: f64,
    pub total_minutes: f64,
    pub unassigned_stops: Vec<Uuid>,
    pub solver_duration_ms: u64,
}
