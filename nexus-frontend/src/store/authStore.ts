import { create } from "zustand";
import { persist } from "zustand/middleware";
import { jwtDecode } from "jwt-decode";

interface JwtPayload {
  sub: string;
  email: string;
  role: string;
  exp: number;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: { id: string; email: string; role: string } | null;
  isAuthenticated: boolean;
  setTokens: (access: string, refresh: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,

      setTokens: (access, refresh) => {
        let user = null;
        try {
          const decoded = jwtDecode<JwtPayload>(access);
          user = { id: decoded.sub, email: decoded.email, role: decoded.role };
        } catch {

        }
        set({ accessToken: access, refreshToken: refresh, user, isAuthenticated: !!user });
      },

      logout: () =>
        set({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false }),
    }),
    {
      name: "nexus-auth",
      partialize: (s) => ({ accessToken: s.accessToken, refreshToken: s.refreshToken }),
    }
  )
);
