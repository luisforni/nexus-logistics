import axios, { AxiosError, AxiosInstance } from "axios";
import { useAuthStore } from "@/store/authStore";

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

function createApiClient(): AxiosInstance {
  const instance = axios.create({
    baseURL: BASE_URL,
    timeout: 15_000,
    headers: { "Content-Type": "application/json" },
  });

  instance.interceptors.request.use((config) => {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (res) => res,
    async (error: AxiosError) => {
      const original = error.config as typeof error.config & { _retry?: boolean };
      if (error.response?.status === 401 && !original?._retry) {
        original._retry = true;
        const refreshToken = useAuthStore.getState().refreshToken;
        if (refreshToken) {
          try {
            const { data } = await axios.post(`${BASE_URL}/auth/refresh`, {
              refresh_token: refreshToken,
            });
            useAuthStore.getState().setTokens(data.access_token, data.refresh_token);
            instance.defaults.headers.common.Authorization = `Bearer ${data.access_token}`;
            return instance(original!);
          } catch {
            useAuthStore.getState().logout();
          }
        } else {
          useAuthStore.getState().logout();
        }
      }
      return Promise.reject(error);
    }
  );

  return instance;
}

export const api = createApiClient();

export interface LoginPayload {
  email: string;
  password: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface Address {
  street: string;
  city: string;
  state: string;
  country: string;
  postal_code: string;
  latitude: number;
  longitude: number;
}

export interface ShipmentEvent {
  id: string;
  shipment_id: string;
  status: string;
  location: Address;
  notes: string;
  recorded_by: string;
  tx_hash: string;
  created_at: string;
}

export interface Shipment {
  id: string;
  tracking_number: string;
  status: string;
  sender_id: string;
  recipient_name: string;
  recipient_email: string;
  origin: Address;
  destination: Address;
  weight_kg: number;
  estimated_at: string;
  delivered_at?: string;
  blockchain_tx_hash?: string;
  events: ShipmentEvent[];
  created_at: string;
  updated_at: string;
}

export interface ListResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface RouteOptimizationRequest {
  depot: { lat: number; lng: number };
  stops: Array<{
    id: string;
    label: string;
    coordinate: { lat: number; lng: number };
    demand: number;
    service_minutes: number;
    time_window?: [number, number];
  }>;
  vehicles: Array<{ id: string; capacity: number; max_stops: number }>;
  objective?: "minimize_distance" | "minimize_time" | "balance_load";
}

export const authApi = {
  login: (payload: LoginPayload) =>
    api.post<TokenPair>("/auth/login", payload).then((r) => r.data),
  refresh: (token: string) =>
    api.post<TokenPair>("/auth/refresh", { refresh_token: token }).then((r) => r.data),
};

export const shipmentsApi = {
  list: (limit = 20, offset = 0) =>
    api.get<ListResponse<Shipment>>("/shipments", { params: { limit, offset } }).then((r) => r.data),
  getById: (id: string) =>
    api.get<Shipment>(`/shipments/${id}`).then((r) => r.data),
  create: (payload: Partial<Shipment>) =>
    api.post<Shipment>("/shipments", payload).then((r) => r.data),
  updateStatus: (id: string, status: string, notes?: string) =>
    api.put<Shipment>(`/shipments/${id}/status`, { status, notes }).then((r) => r.data),
  getTrace: (id: string) =>
    api.get(`/shipments/${id}/trace`).then((r) => r.data),
};

export const analyticsApi = {
  kpis: () => api.get("/analytics/kpis").then((r) => r.data),
  forecast: () => api.get("/analytics/forecast").then((r) => r.data),
};

export const optimizerApi = {
  optimizeRoute: (req: RouteOptimizationRequest) =>
    api.post("/optimize/route", req).then((r) => r.data),
};
