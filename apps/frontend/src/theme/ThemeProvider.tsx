import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api, type TenantPublicConfig } from "../api/client";

interface ThemeContextValue {
  config: TenantPublicConfig | null;
  loading: boolean;
}

const ThemeContext = createContext<ThemeContextValue>({ config: null, loading: true });

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [config, setConfig] = useState<TenantPublicConfig | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    api
      .getPublicConfig()
      .then((cfg) => {
        if (!mounted) return;
        setConfig(cfg);

        const root = document.documentElement;
        root.style.setProperty("--primary-color", cfg.branding?.primary_color ?? "#1f3b5c");
        root.style.setProperty("--secondary-color", cfg.branding?.secondary_color ?? "#d6b370");
      })
      .catch(() => {
        const root = document.documentElement;
        root.style.setProperty("--primary-color", "#1f3b5c");
        root.style.setProperty("--secondary-color", "#d6b370");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });

    return () => {
      mounted = false;
    };
  }, []);

  const value = useMemo(() => ({ config, loading }), [config, loading]);
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  return useContext(ThemeContext);
}
