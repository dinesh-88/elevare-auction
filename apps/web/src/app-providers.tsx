"use client";

import LocaleProvider from "./components/LocaleProvider";
import { AuthProvider } from "./hooks/useAuth";
import { ThemeProvider } from "./theme/ThemeProvider";

export default function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <LocaleProvider>
      <AuthProvider>
        <ThemeProvider>{children}</ThemeProvider>
      </AuthProvider>
    </LocaleProvider>
  );
}
