import { useQuery } from "@tanstack/react-query";
import { analyticsApi, shipmentsApi } from "@/api/client";
import { Package, TrendingUp, Truck, CheckCircle2, AlertCircle, Clock } from "lucide-react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import ShipmentCard from "@/components/ShipmentCard";

const CHART_DATA = Array.from({ length: 14 }, (_, i) => ({
  day: `Mar ${i + 1}`,
  shipments: Math.round(180 + Math.random() * 60 - 30),
  delivered: Math.round(160 + Math.random() * 40 - 20),
}));

interface KPIs {
  total_shipments: number;
  on_time_rate: number;
  avg_delivery_hours: number;
}

export default function Dashboard() {
  const { data: kpis } = useQuery<KPIs>({
    queryKey: ["kpis"],
    queryFn: analyticsApi.kpis,
  });

  const { data: recent } = useQuery({
    queryKey: ["shipments", "recent"],
    queryFn: () => shipmentsApi.list(6, 0),
  });

  const stats = [
    {
      label: "Total Shipments",
      value: kpis?.total_shipments?.toLocaleString() ?? "—",
      icon: Package,
      color: "text-blue-400",
    },
    {
      label: "On-Time Rate",
      value: kpis ? `${(kpis.on_time_rate * 100).toFixed(1)}%` : "—",
      icon: TrendingUp,
      color: "text-green-400",
    },
    {
      label: "Avg Delivery",
      value: kpis ? `${kpis.avg_delivery_hours}h` : "—",
      icon: Clock,
      color: "text-yellow-400",
    },
    {
      label: "In Transit",
      value: "—",
      icon: Truck,
      color: "text-indigo-400",
    },
  ];

  return (
    <div className="p-6 space-y-6">
      {}
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-0.5">Real-time supply chain overview</p>
      </div>

      {}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="card flex items-start gap-4">
            <div className={`rounded-lg bg-gray-800 p-2 ${color}`}>
              <Icon className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs text-gray-500">{label}</p>
              <p className="text-2xl font-bold mt-0.5">{value}</p>
            </div>
          </div>
        ))}
      </div>

      {}
      <div className="card">
        <h2 className="mb-4 text-sm font-semibold text-gray-300">Shipment Volume — Last 14 days</h2>
        <ResponsiveContainer width="100%" height={220}>
          <AreaChart data={CHART_DATA} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="shipGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="delGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
            <XAxis dataKey="day" tick={{ fill: "#6b7280", fontSize: 11 }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fill: "#6b7280", fontSize: 11 }} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ background: "#111827", border: "1px solid #374151", borderRadius: 8 }}
              labelStyle={{ color: "#d1d5db" }}
              itemStyle={{ color: "#9ca3af" }}
            />
            <Area type="monotone" dataKey="shipments" stroke="#3b82f6" fill="url(#shipGrad)" name="Created" strokeWidth={2} dot={false} />
            <Area type="monotone" dataKey="delivered" stroke="#22c55e" fill="url(#delGrad)"  name="Delivered" strokeWidth={2} dot={false} />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {}
      <div>
        <h2 className="mb-4 text-sm font-semibold text-gray-300">Recent Shipments</h2>
        {recent?.data?.length ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {recent.data.map((s) => (
              <ShipmentCard key={s.id} shipment={s} />
            ))}
          </div>
        ) : (
          <div className="card flex flex-col items-center gap-3 py-12 text-center">
            <Package className="h-10 w-10 text-gray-700" />
            <p className="text-sm text-gray-500">No shipments yet.</p>
          </div>
        )}
      </div>
    </div>
  );
}
