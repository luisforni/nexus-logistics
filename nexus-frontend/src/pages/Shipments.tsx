import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { shipmentsApi } from "@/api/client";
import ShipmentCard from "@/components/ShipmentCard";
import { Search, Package, ChevronLeft, ChevronRight, Loader2 } from "lucide-react";
import { useShipmentStore } from "@/store/shipmentStore";

const PAGE_SIZE = 12;

export default function Shipments() {
  const [page, setPage] = useState(0);
  const { filter, setFilter } = useShipmentStore();

  const { data, isLoading, isError } = useQuery({
    queryKey: ["shipments", page],
    queryFn: () => shipmentsApi.list(PAGE_SIZE, page * PAGE_SIZE),
    placeholderData: (prev) => prev,
  });

  const filtered = (data?.data ?? []).filter(
    (s) =>
      !filter ||
      s.tracking_number.toLowerCase().includes(filter.toLowerCase()) ||
      s.recipient_name.toLowerCase().includes(filter.toLowerCase())
  );

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;

  return (
    <div className="p-6 space-y-5">
      {}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Shipments</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            {data ? `${data.total.toLocaleString()} total` : "Loading…"}
          </p>
        </div>
      </div>

      {}
      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500 pointer-events-none" />
        <input
          className="input pl-9"
          placeholder="Search tracking # or recipient…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      {}
      {isLoading ? (
        <div className="flex items-center gap-2 text-gray-500 py-20 justify-center">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>Loading shipments…</span>
        </div>
      ) : isError ? (
        <div className="card flex flex-col items-center gap-3 py-16 text-center">
          <Package className="h-10 w-10 text-red-800" />
          <p className="text-sm text-red-400">Failed to load shipments.</p>
        </div>
      ) : filtered.length === 0 ? (
        <div className="card flex flex-col items-center gap-3 py-16 text-center">
          <Package className="h-10 w-10 text-gray-700" />
          <p className="text-sm text-gray-500">No shipments match your search.</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filtered.map((s) => (
            <ShipmentCard key={s.id} shipment={s} />
          ))}
        </div>
      )}

      {}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-2">
          <button
            className="btn-ghost"
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
          >
            <ChevronLeft className="h-4 w-4" />
            Prev
          </button>
          <span className="text-sm text-gray-500">
            Page {page + 1} / {totalPages}
          </span>
          <button
            className="btn-ghost"
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  );
}
