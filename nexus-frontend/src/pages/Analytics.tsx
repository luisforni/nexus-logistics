import { useQuery } from "@tanstack/react-query";
import { analyticsApi } from "@/api/client";
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";

const DEMAND_DATA = Array.from({ length: 30 }, (_, i) => {
  const isPast = i < 20;
  return {
    day: `D${i + 1}`,
    actual: isPast ? Math.round(200 + Math.random() * 80 - 40) : undefined,
    forecast: !isPast ? Math.round(220 + Math.random() * 60 - 30) : undefined,
    lower: !isPast ? Math.round(180 + Math.random() * 40) : undefined,
    upper: !isPast ? Math.round(260 + Math.random() * 40) : undefined,
  };
});

const STATUS_DIST = [
  { status: "Delivered",   count: 2841 },
  { status: "In Transit",  count: 412 },
  { status: "Pending",     count: 288 },
  { status: "At Hub",      count: 143 },
  { status: "Out f. Del.", count: 97 },
  { status: "Failed",      count: 40 },
];

export default function Analytics() {
  const { data: kpis } = useQuery({
    queryKey: ["kpis"],
    queryFn: analyticsApi.kpis,
  });

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Analytics</h1>
        <p className="text-sm text-gray-500 mt-0.5">AI-driven forecasting & operational KPIs</p>
      </div>

      {}
      <div className="card">
        <h2 className="mb-1 text-sm font-semibold text-gray-300">Demand Forecast (Prophet model)</h2>
        <p className="text-xs text-gray-600 mb-4">Historical + 10-day forecast with 95% confidence intervals</p>
        <ResponsiveContainer width="100%" height={260}>
          <LineChart data={DEMAND_DATA} margin={{ left: -20, right: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
            <XAxis dataKey="day" tick={{ fill: "#6b7280", fontSize: 10 }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fill: "#6b7280", fontSize: 10 }} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ background: "#111827", border: "1px solid #374151", borderRadius: 8 }}
              itemStyle={{ color: "#9ca3af" }}
            />
            <Legend wrapperStyle={{ fontSize: 12, color: "#6b7280" }} />
            <Line type="monotone" dataKey="actual"   stroke="#3b82f6" strokeWidth={2} dot={false} name="Actual" connectNulls />
            <Line type="monotone" dataKey="forecast" stroke="#f59e0b" strokeWidth={2} dot={false} name="Forecast" strokeDasharray="5 5" connectNulls />
            <Line type="monotone" dataKey="upper"    stroke="#f59e0b" strokeWidth={0.5} dot={false} name="Upper CI" strokeDasharray="2 4" connectNulls />
            <Line type="monotone" dataKey="lower"    stroke="#f59e0b" strokeWidth={0.5} dot={false} name="Lower CI" strokeDasharray="2 4" connectNulls />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {}
      <div className="card">
        <h2 className="mb-4 text-sm font-semibold text-gray-300">Shipment Status Distribution</h2>
        <ResponsiveContainer width="100%" height={220}>
          <BarChart data={STATUS_DIST} margin={{ left: -20, right: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
            <XAxis dataKey="status" tick={{ fill: "#6b7280", fontSize: 10 }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fill: "#6b7280", fontSize: 10 }} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ background: "#111827", border: "1px solid #374151", borderRadius: 8 }}
              itemStyle={{ color: "#9ca3af" }}
            />
            <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} name="Count" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {}
      {kpis && (
        <div className="grid sm:grid-cols-3 gap-4">
          {Object.entries(kpis).map(([k, v]) => (
            <div key={k} className="card">
              <p className="text-xs text-gray-500 mb-1 capitalize">{k.replace(/_/g, " ")}</p>
              <p className="text-2xl font-bold">{String(v)}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
