import { Outlet, NavLink, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Package,
  BarChart3,
  LogOut,
  Boxes,
} from "lucide-react";
import { useAuthStore } from "@/store/authStore";

const nav = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard, end: true },
  { to: "/shipments", label: "Shipments", icon: Package },
  { to: "/analytics", label: "Analytics", icon: BarChart3 },
];

export default function Layout() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {}
      <aside className="flex w-60 flex-col bg-gray-900 border-r border-gray-800">
        {}
        <div className="flex items-center gap-2.5 px-5 py-5 border-b border-gray-800">
          <Boxes className="h-7 w-7 text-blue-400" />
          <span className="text-lg font-bold tracking-tight">Nexus</span>
          <span className="ml-auto text-xs text-gray-600 font-mono">v1.0</span>
        </div>

        {}
        <nav className="flex-1 space-y-0.5 px-3 py-4">
          {nav.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                  isActive
                    ? "bg-blue-600 text-white"
                    : "text-gray-400 hover:bg-gray-800 hover:text-gray-100"
                }`
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              {label}
            </NavLink>
          ))}
        </nav>

        {}
        <div className="border-t border-gray-800 p-4">
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center text-xs font-bold uppercase">
              {user?.email?.charAt(0) ?? "U"}
            </div>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-medium text-gray-200">{user?.email}</p>
              <p className="text-xs text-gray-500 capitalize">{user?.role?.toLowerCase()}</p>
            </div>
            <button
              onClick={handleLogout}
              className="text-gray-500 hover:text-red-400 transition"
              title="Logout"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {}
      <main className="flex-1 overflow-y-auto bg-gray-950">
        <Outlet />
      </main>
    </div>
  );
}
