"use client";

import { AuthProvider } from "./hooks/useAuth";
import { ThemeProvider } from "./theme/ThemeProvider";

export default function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <ThemeProvider>{children}</ThemeProvider>
    </AuthProvider>
  );
}
