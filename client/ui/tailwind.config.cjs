/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{svelte,js,ts}"],
  theme: {
    extend: {
      colors: {
        shell: {
          bg: "#f2f7f6",
          ink: "#13252b",
          panel: "#ffffff",
          line: "#d4e4e8",
          muted: "#55727a",
          accent: "#0ea5a1",
          success: "#22c55e"
        }
      },
      boxShadow: {
        shell: "0 20px 40px rgba(16, 46, 56, 0.12)"
      },
      fontFamily: {
        display: ["Space Grotesk", "Trebuchet MS", "sans-serif"],
        mono: ["IBM Plex Mono", "monospace"]
      }
    }
  },
  plugins: []
};
