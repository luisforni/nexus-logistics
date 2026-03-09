import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { shipmentsApi } from "@/api/client";
import StatusBadge from "@/components/StatusBadge";
import { ArrowLeft, Link2, MapPin, Weight, Calendar, User } from "lucide-react";
import { format } from "date-fns";

function Field({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <p className="text-xs text-gray-500 mb-0.5">{label}</p>
      <p className="text-sm text-gray-200">{value ?? "—"}</p>
    </div>
  );
}

export default function ShipmentDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: shipment, isLoading, isError } = useQuery({
    queryKey: ["shipment", id],
    queryFn: () => shipmentsApi.getById(id!),
    enabled: !!id,
  });

  if (isLoading) return <div className="p-6 text-gray-500">Loading…</div>;
  if (isError || !shipment)
    return <div className="p-6 text-red-400">Shipment not found.</div>;

  return (
    <div className="p-6 space-y-6 max-w-4xl">
      {}
      <button onClick={() => navigate(-1)} className="btn-ghost -ml-2">
        <ArrowLeft className="h-4 w-4" />
        Back
      </button>

      {}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold font-mono">{shipment.tracking_number}</h1>
          <p className="text-sm text-gray-500 mt-0.5">ID: {shipment.id}</p>
        </div>
        <StatusBadge status={shipment.status} />
      </div>

      {}
      <div className="grid sm:grid-cols-2 gap-4">
        <div className="card space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-500">Shipment Info</h2>
          <Field label="Recipient" value={<span className="flex items-center gap-1.5"><User className="h-3.5 w-3.5 text-gray-500" />{shipment.recipient_name}</span>} />
          <Field label="Recipient Email" value={shipment.recipient_email} />
          <Field label="Weight" value={<span className="flex items-center gap-1.5"><Weight className="h-3.5 w-3.5 text-gray-500" />{shipment.weight_kg} kg</span>} />
          <Field label="Estimated Delivery" value={<span className="flex items-center gap-1.5"><Calendar className="h-3.5 w-3.5 text-gray-500" />{format(new Date(shipment.estimated_at), "PPP")}</span>} />
          {shipment.delivered_at && (
            <Field label="Delivered At" value={format(new Date(shipment.delivered_at), "PPP HH:mm")} />
          )}
        </div>

        <div className="card space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-500">Route</h2>
          <Field
            label="Origin"
            value={
              <span className="flex items-start gap-1.5">
                <MapPin className="h-3.5 w-3.5 text-blue-400 mt-0.5 shrink-0" />
                <span>{shipment.origin?.street}, {shipment.origin?.city}, {shipment.origin?.country}</span>
              </span>
            }
          />
          <Field
            label="Destination"
            value={
              <span className="flex items-start gap-1.5">
                <MapPin className="h-3.5 w-3.5 text-green-400 mt-0.5 shrink-0" />
                <span>{shipment.destination?.street}, {shipment.destination?.city}, {shipment.destination?.country}</span>
              </span>
            }
          />
        </div>
      </div>

      {}
      {shipment.blockchain_tx_hash && (
        <div className="card flex items-center gap-3">
          <Link2 className="h-4 w-4 text-green-400 shrink-0" />
          <div className="min-w-0">
            <p className="text-xs text-gray-500 mb-0.5">On-Chain Anchor (Ethereum)</p>
            <p className="text-sm font-mono text-green-300 truncate">{shipment.blockchain_tx_hash}</p>
          </div>
          <span className="ml-auto h-2 w-2 rounded-full bg-green-400 animate-pulse shrink-0" />
        </div>
      )}

      {}
      <div className="card">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-500 mb-4">
          Immutable Event Timeline
        </h2>
        {shipment.events?.length ? (
          <ol className="relative border-l border-gray-800 ml-1 space-y-6">
            {[...shipment.events].reverse().map((evt) => (
              <li key={evt.id} className="ml-5">
                <span className="absolute -left-[7px] mt-1 h-3.5 w-3.5 rounded-full bg-blue-600 border-2 border-gray-950" />
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-0.5">
                    <StatusBadge status={evt.status} size="sm" />
                    {evt.notes && (
                      <p className="text-xs text-gray-400 mt-1">{evt.notes}</p>
                    )}
                    {evt.location?.city && (
                      <p className="text-xs text-gray-600 flex items-center gap-1 mt-0.5">
                        <MapPin className="h-3 w-3" />
                        {evt.location.city}, {evt.location.country}
                      </p>
                    )}
                    {evt.tx_hash && (
                      <p className="text-[10px] font-mono text-gray-600 mt-0.5">
                        tx: {evt.tx_hash.slice(0, 22)}…
                      </p>
                    )}
                  </div>
                  <time className="text-xs text-gray-600 shrink-0">
                    {format(new Date(evt.created_at), "MMM d HH:mm")}
                  </time>
                </div>
              </li>
            ))}
          </ol>
        ) : (
          <p className="text-sm text-gray-600">No events recorded yet.</p>
        )}
      </div>
    </div>
  );
}
