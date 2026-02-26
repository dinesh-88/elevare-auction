import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{js,ts,jsx,tsx}", "./src/**/*.{js,ts,jsx,tsx}", "./components/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        heading: ['"Playfair Display"', "serif"],
        body: ["Inter", "sans-serif"]
      },
      colors: {
        brand: {
          900: "#1C1C1C",
          800: "#222222",
          700: "#2E2E2E",
          primary: "#1C1C1C",
          secondary: "#222222",
          accent: "#9E874C"
        },
        accent: {
          600: "#9E874C",
          500: "#B79B57",
          400: "#C9B67A"
        },
        neutral: {
          50: "#F8F6F2",
          100: "#F1EFEA",
          200: "#E5E2DB",
          300: "#BDBDB6",
          400: "#9E9E97",
          500: "#7F7F78",
          600: "#5F5F59",
          700: "#4A4A4A",
          800: "#2E2E2E",
          900: "#1C1C1C"
        },
        success: {
          500: "#2E5F4F"
        },
        danger: {
          500: "#7A2E2E"
        }
      }
    }
  },
  plugins: []
};

export default config;
