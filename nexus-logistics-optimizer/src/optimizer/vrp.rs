

use std::time::Instant;
use uuid::Uuid;

use crate::models::{Coordinate, OptimizationResult, Route, RouteOptimizationRequest, Stop, Vehicle};

pub fn solve(req: &RouteOptimizationRequest) -> OptimizationResult {
    let start = Instant::now();

    if req.stops.is_empty() || req.vehicles.is_empty() {
        return OptimizationResult {
            routes: vec![],
            total_distance_km: 0.0,
            total_minutes: 0.0,
            unassigned_stops: vec![],
            solver_duration_ms: 0,
        };
    }

    let mut savings = compute_savings(&req.depot, &req.stops);
    savings.sort_by(|(_, _, a), (_, _, b)| b.partial_cmp(a).unwrap());

    let mut routes = construct_routes(&req.stops, &req.vehicles, &savings, &req.depot);

    for r in &mut routes {
        two_opt_improve(r, &req.stops, &req.depot);
    }

    let assigned: std::collections::HashSet<Uuid> = routes
        .iter()
        .flat_map(|r| r.stop_sequence.iter().copied())
        .collect();

    let unassigned = req.stops
        .iter()
        .filter(|s| !assigned.contains(&s.id))
        .map(|s| s.id)
        .collect::<Vec<_>>();

    let total_distance: f64 = routes.iter().map(|r| r.total_distance_km).sum();
    let total_minutes: f64 = routes.iter().map(|r| r.total_minutes).sum();

    OptimizationResult {
        routes,
        total_distance_km: total_distance,
        total_minutes,
        unassigned_stops: unassigned,
        solver_duration_ms: start.elapsed().as_millis() as u64,
    }
}

fn compute_savings(depot: &Coordinate, stops: &[Stop]) -> Vec<(usize, usize, f64)> {
    let n = stops.len();
    let mut savings = Vec::with_capacity(n * n / 2);
    for i in 0..n {
        for j in (i + 1)..n {
            let d_depot_i = haversine(depot, &stops[i].coordinate);
            let d_depot_j = haversine(depot, &stops[j].coordinate);
            let d_ij      = haversine(&stops[i].coordinate, &stops[j].coordinate);
            let saving = d_depot_i + d_depot_j - d_ij;
            savings.push((i, j, saving));
        }
    }
    savings
}

fn construct_routes(
    stops: &[Stop],
    vehicles: &[Vehicle],
    savings: &[(usize, usize, f64)],
    depot: &Coordinate,
) -> Vec<Route> {

    let mut route_of: Vec<Option<usize>> = vec![None; stops.len()];
    let mut routes: Vec<Vec<usize>> = (0..stops.len()).map(|i| vec![i]).collect();
    let mut loads: Vec<f64> = stops.iter().map(|s| s.demand).collect();

    for &(i, j, _saving) in savings {
        let ri = match route_of[i] { Some(r) => r, None => i };
        let rj = match route_of[j] { Some(r) => r, None => j };
        if ri == rj { continue; }

        let combined_load = loads[ri] + loads[rj];

        let fits = vehicles.iter().any(|v| {
            combined_load <= v.capacity
                && routes[ri].len() + routes[rj].len() <= v.max_stops
        });

        if fits {
            let merged: Vec<usize> = routes[ri].iter().chain(routes[rj].iter()).copied().collect();
            let new_load = combined_load;
            routes[ri] = merged.clone();
            loads[ri] = new_load;
            routes[rj] = vec![];
            loads[rj] = 0.0;
            for &s in &merged {
                route_of[s] = Some(ri);
            }
        }
    }

    let non_empty: Vec<Vec<usize>> = routes.into_iter().filter(|r| !r.is_empty()).collect();

    non_empty
        .into_iter()
        .zip(vehicles.iter().cycle())
        .map(|(stop_indices, vehicle)| {
            let stop_ids: Vec<Uuid> = stop_indices.iter().map(|&i| stops[i].id).collect();
            let (dist, mins) = route_metrics(depot, &stop_indices, stops);
            Route {
                vehicle_id: vehicle.id,
                stop_sequence: stop_ids,
                total_distance_km: dist,
                total_minutes: mins,
                total_demand: stop_indices.iter().map(|&i| stops[i].demand).sum(),
            }
        })
        .collect()
}

fn two_opt_improve(route: &mut Route, stops: &[Stop], depot: &Coordinate) {
    let n = route.stop_sequence.len();
    if n < 4 { return; }

    let id_to_index: std::collections::HashMap<Uuid, usize> = stops
        .iter()
        .enumerate()
        .map(|(i, s)| (s.id, i))
        .collect();

    let indices: Vec<usize> = route.stop_sequence
        .iter()
        .filter_map(|id| id_to_index.get(id).copied())
        .collect();

    let mut best = indices.clone();
    let mut improved = true;

    while improved {
        improved = false;
        for i in 0..(best.len() - 1) {
            for j in (i + 2)..best.len() {
                let mut candidate = best.clone();
                candidate[i..=j].reverse();

                let (old_d, _) = route_metrics(depot, &best, stops);
                let (new_d, _) = route_metrics(depot, &candidate, stops);
                if new_d < old_d - 1e-9 {
                    best = candidate;
                    improved = true;
                }
            }
        }
    }

    route.stop_sequence = best.iter().map(|&i| stops[i].id).collect();
    let (dist, mins) = route_metrics(depot, &best, stops);
    route.total_distance_km = dist;
    route.total_minutes = mins;
}

fn route_metrics(depot: &Coordinate, indices: &[usize], stops: &[Stop]) -> (f64, f64) {
    if indices.is_empty() {
        return (0.0, 0.0);
    }
    let mut dist = haversine(depot, &stops[indices[0]].coordinate);
    for w in indices.windows(2) {
        dist += haversine(&stops[w[0]].coordinate, &stops[w[1]].coordinate);
    }
    dist += haversine(&stops[*indices.last().unwrap()].coordinate, depot);

    let service: f64 = indices.iter().map(|&i| stops[i].service_minutes as f64).sum();

    let drive_minutes = (dist / 50.0) * 60.0;
    (dist, drive_minutes + service)
}

pub fn haversine(a: &Coordinate, b: &Coordinate) -> f64 {
    const R: f64 = 6371.0;
    let lat1 = a.lat.to_radians();
    let lat2 = b.lat.to_radians();
    let dlat = (b.lat - a.lat).to_radians();
    let dlng = (b.lng - a.lng).to_radians();
    let h = (dlat / 2.0).sin().powi(2)
        + lat1.cos() * lat2.cos() * (dlng / 2.0).sin().powi(2);
    2.0 * R * h.sqrt().asin()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_stop(lat: f64, lng: f64, demand: f64) -> Stop {
        Stop {
            id: Uuid::new_v4(),
            label: "test".into(),
            coordinate: Coordinate { lat, lng },
            time_window: None,
            service_minutes: 10,
            demand,
        }
    }

    #[test]
    fn haversine_same_point_is_zero() {
        let c = Coordinate { lat: 51.5074, lng: -0.1278 };
        assert!((haversine(&c, &c)).abs() < 1e-9);
    }

    #[test]
    fn haversine_london_paris_approx() {
        let london = Coordinate { lat: 51.5074, lng: -0.1278 };
        let paris  = Coordinate { lat: 48.8566, lng:  2.3522 };
        let d = haversine(&london, &paris);
        assert!((d - 340.0).abs() < 20.0, "expected ~340 km, got {d}");
    }

    #[test]
    fn solve_empty_request_returns_empty() {
        let req = RouteOptimizationRequest {
            depot: Coordinate { lat: 0.0, lng: 0.0 },
            stops: vec![],
            vehicles: vec![],
            objective: Default::default(),
        };
        let result = solve(&req);
        assert!(result.routes.is_empty());
    }

    #[test]
    fn solve_assigns_stops_to_vehicles() {
        let depot = Coordinate { lat: 40.7128, lng: -74.0060 };
        let stops = vec![
            make_stop(40.73, -73.99, 10.0),
            make_stop(40.75, -74.01, 15.0),
            make_stop(40.71, -74.03, 20.0),
        ];
        let vehicles = vec![Vehicle {
            id: Uuid::new_v4(),
            capacity: 100.0,
            max_stops: 10,
        }];
        let req = RouteOptimizationRequest { depot, stops, vehicles, objective: Default::default() };
        let result = solve(&req);
        let assigned: usize = result.routes.iter().map(|r| r.stop_sequence.len()).sum();
        assert_eq!(assigned + result.unassigned_stops.len(), 3);
    }
}
