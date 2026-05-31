/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{svelte,js,ts}"],
  theme: {
    extend: {
      colors: {
        // SecureCollab Design System Colors
        ivory: "#FAF9F6",        // Main background - warm off-white
        sidebar: "#F5F2ED",      // Sidebar background - slightly warmer
        sage: "#8A9A5B",         // Primary accent - calm green
        clay: "#B07D62",         // Secondary accent - warm terracotta
        charcoal: "#2C2C2C",     // Primary text - dark charcoal
        muted: "#7A7672",        // Secondary text - gray-brown
        borderSoft: "#EAE3D9",   // Borders - very light
        stone: {
          200: "#e5e7eb",
          300: "#d1d5db",
          400: "#9ca3af",
          500: "#6b7280",
          800: "#1f2937",
        }
      },
      fontFamily: {
        sans: ['"Outfit"', '"Inter"', '"Segoe UI"', '"system-ui"', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', '"IBM Plex Mono"', 'monospace'],
      },
      fontSize: {
        "2xs": ["0.625rem", { lineHeight: "0.875rem" }],
        "xs": ["12px", { lineHeight: "1.4" }],
        "sm": ["13px", { lineHeight: "1.4" }],
        "base": ["14px", { lineHeight: "1.6" }],
        "lg": ["15px", { lineHeight: "1.3" }],
        "xl": ["18px", { lineHeight: "1.2" }],
      },
      boxShadow: {
        sm: "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
        md: "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)",
        lg: "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)",
        xl: "0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)",
      },
      keyframes: {
        "fade-in": { from: { opacity: "0" }, to: { opacity: "1" } },
        "slide-up": { from: { opacity: "0", transform: "translateY(4px)" }, to: { opacity: "1", transform: "translateY(0)" } },
      },
      animation: {
        "fade-in":  "fade-in 150ms ease",
        "slide-up": "slide-up 150ms ease",
      },
    },
  },
  plugins: [],
};
