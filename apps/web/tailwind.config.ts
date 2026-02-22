import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{js,ts,jsx,tsx}", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        heading: ['"Playfair Display"', "serif"],
        body: ["Inter", "sans-serif"]
      },
      colors: {
        brand: {
          primary: "var(--brand-primary)",
          secondary: "var(--brand-secondary)",
          accent: "var(--brand-accent)"
        },
        neutral: {
          50: "#F7F7F5",
          100: "#ECECE8",
          200: "#DCDCD6",
          300: "#BDBDB6",
          400: "#9E9E97",
          500: "#7F7F78",
          600: "#5F5F59",
          700: "#3F3F3A",
          800: "#2A2A26",
          900: "#1C1C1A"
        }
      }
    }
  },
  plugins: []
};

export default config;
