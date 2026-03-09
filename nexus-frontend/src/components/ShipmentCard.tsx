import { Link } from "react-router-dom";
import { Package, MapPin, Calendar } from "lucide-react";
import { Shipment } from "@/api/client";
import StatusBadge from "./StatusBadge";
import { format } from "date-fns";

interface Props {
  shipment: Shipment;
}

export default function ShipmentCard({ shipment }: Props) {
  return (
    <Link
      to={`/shipments/${shipment.id}`}
      className="card flex flex-col gap-3 hover:border-blue-700 transition-colors cursor-pointer"
    >
      {}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <Package className="h-4 w-4 text-blue-400 shrink-0" />
          <span className="font-mono text-sm font-semibold text-blue-300">
            {shipment.tracking_number}
          </span>
        </div>
        <StatusBadge status={shipment.status} size="sm" />
      </div>

      {}
      <p className="text-sm font-medium text-gray-200 truncate">{shipment.recipient_name}</p>

      {}
      <div className="flex items-center gap-1.5 text-xs text-gray-500">
        <MapPin className="h-3.5 w-3.5 shrink-0" />
        <span className="truncate">
          {shipment.origin?.city}, {shipment.origin?.country}
        </span>
        <span className="text-gray-700 mx-0.5">→</span>
        <span className="truncate">
          {shipment.destination?.city}, {shipment.destination?.country}
        </span>
      </div>

      {}
      <div className="flex items-center justify-between text-xs text-gray-600">
        <span className="flex items-center gap-1">
          <Calendar className="h-3.5 w-3.5" />
          ETA: {format(new Date(shipment.estimated_at), "MMM d, yyyy")}
        </span>
        <span>{shipment.weight_kg} kg</span>
      </div>

      {}
      {shipment.blockchain_tx_hash && (
        <div className="flex items-center gap-1.5 rounded-md bg-gray-800 px-2 py-1">
          <span className="h-1.5 w-1.5 rounded-full bg-green-400 animate-pulse" />
          <span className="text-[10px] text-gray-400 font-mono truncate">
            on-chain: {shipment.blockchain_tx_hash.slice(0, 18)}…
          </span>
        </div>
      )}
    </Link>
  );
}
