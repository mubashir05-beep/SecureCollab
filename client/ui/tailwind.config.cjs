/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{svelte,js,ts}"],
  theme: {
    extend: {
      colors: {
        // Dark shell — Slack-inspired dark palette
        shell: {
          // Sidebar / workspace rail
          sidebar:   "#1a1d21",  // deepest dark — workspace rail bg
          panel:     "#19171d",  // channel panel bg (slightly purple-tinted dark)
          // Main content
          bg:        "#1d2026",  // main area background
          surface:   "#222529",  // message hover / elevated surfaces
          elevated:  "#2b2d31",  // cards, tooltips, modals within dark
          // Borders
          border:    "#3f4147",  // default divider
          borderSub: "#2b2d31",  // subtle divider
          // Text
          ink:       "#e3e5e8",  // primary text
          muted:     "#949ba4",  // secondary / metadata text
          subtle:    "#6d7379",  // placeholder / disabled text
          // Accent — indigo
          accent:    "#5865f2",  // Discord/Notion-style indigo — primary CTA
          accentHov: "#4752c4",  // hover
          accentText:"#c9ccff",  // accent on dark bg
          // Semantic
          success:   "#23a55a",
          warn:      "#f0b232",
          danger:    "#f23f42",
          dangerBg:  "#3a1b1b",
          // Mention highlight
          mention:   "#444271",
          mentionTxt:"#c9ccff",
        }
      },
      fontFamily: {
        sans:    ['"Inter"', '"Segoe UI"', '"system-ui"', 'sans-serif'],
        mono:    ['"JetBrains Mono"', '"Fira Code"', '"IBM Plex Mono"', 'monospace'],
      },
      fontSize: {
        "2xs": ["0.625rem", { lineHeight: "0.875rem" }],
      },
      boxShadow: {
        panel:  "0 2px 8px rgba(0,0,0,0.4)",
        modal:  "0 8px 32px rgba(0,0,0,0.6)",
        tooltip:"0 2px 4px rgba(0,0,0,0.5)",
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
