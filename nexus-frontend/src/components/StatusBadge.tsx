const STATUS_STYLES: Record<string, string> = {
  PENDING:          "bg-gray-700  text-gray-300",
  PICKED_UP:        "bg-blue-900  text-blue-300",
  IN_TRANSIT:       "bg-indigo-900 text-indigo-300",
  AT_HUB:           "bg-purple-900 text-purple-300",
  OUT_FOR_DELIVERY: "bg-yellow-900 text-yellow-300",
  DELIVERED:        "bg-green-900  text-green-300",
  FAILED:           "bg-red-900    text-red-300",
  RETURNED:         "bg-orange-900 text-orange-300",
};

const STATUS_LABELS: Record<string, string> = {
  PENDING:          "Pending",
  PICKED_UP:        "Picked Up",
  IN_TRANSIT:       "In Transit",
  AT_HUB:           "At Hub",
  OUT_FOR_DELIVERY: "Out for Delivery",
  DELIVERED:        "Delivered",
  FAILED:           "Failed",
  RETURNED:         "Returned",
};

interface Props {
  status: string;
  size?: "sm" | "md";
}

export default function StatusBadge({ status, size = "md" }: Props) {
  const cls = STATUS_STYLES[status] ?? "bg-gray-700 text-gray-300";
  const label = STATUS_LABELS[status] ?? status;
  return (
    <span
      className={`badge ${cls} ${size === "sm" ? "text-[10px] px-2 py-0.5" : ""}`}
    >
      {label}
    </span>
  );
}
