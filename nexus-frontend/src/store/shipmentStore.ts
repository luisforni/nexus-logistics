import { create } from "zustand";
import { Shipment } from "@/api/client";

interface ShipmentState {
  selected: Shipment | null;
  select: (s: Shipment | null) => void;
  filter: string;
  setFilter: (f: string) => void;
}

export const useShipmentStore = create<ShipmentState>()((set) => ({
  selected: null,
  select: (s) => set({ selected: s }),
  filter: "",
  setFilter: (f) => set({ filter: f }),
}));
