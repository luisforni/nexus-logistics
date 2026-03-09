import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { authApi } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import { Boxes, Loader2 } from "lucide-react";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const setTokens = useAuthStore((s) => s.setTokens);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const tokens = await authApi.login({ email, password });
      setTokens(tokens.access_token, tokens.refresh_token);
      navigate("/");
    } catch {
      setError("Invalid email or password.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-950 px-4">
      <div className="w-full max-w-sm">
        {}
        <div className="mb-8 flex flex-col items-center gap-2">
          <Boxes className="h-10 w-10 text-blue-400" />
          <h1 className="text-2xl font-bold">Nexus Logistics</h1>
          <p className="text-sm text-gray-500">Supply Chain Intelligence Platform</p>
        </div>

        {}
        <form onSubmit={handleSubmit} className="card space-y-4">
          <div>
            <label className="mb-1.5 block text-xs font-medium text-gray-400">
              Email address
            </label>
            <input
              type="email"
              className="input"
              placeholder="operator@nexus.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-xs font-medium text-gray-400">
              Password
            </label>
            <input
              type="password"
              className="input"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              autoComplete="current-password"
            />
          </div>

          {error && (
            <p className="rounded-lg bg-red-950 border border-red-800 px-3 py-2 text-xs text-red-400">
              {error}
            </p>
          )}

          <button type="submit" className="btn-primary w-full justify-center" disabled={loading}>
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
            {loading ? "Signing in…" : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}
