export const designTokens = {
  colors: {
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
      900: "#1C1C1A",
    },
  },
  typography: {
    heading: '"Playfair Display", serif',
    body: 'Inter, system-ui, sans-serif',
  },
  spacing: {
    1: "4px",
    2: "8px",
    3: "12px",
    4: "16px",
    5: "20px",
    6: "24px",
    8: "32px",
    10: "40px",
    12: "48px",
    16: "64px",
  },
  layout: {
    containerMaxWidth: "1200px",
    containerPaddingX: "24px",
  },
} as const;
