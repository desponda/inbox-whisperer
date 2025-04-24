/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // Enable class-based dark mode
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        // Windsurf-inspired palette (customize as needed)
        "windsurf-bg": "#1c2431",
        "windsurf-accent": "#38bdf8",
        "windsurf-card": "#232b3b",
        "windsurf-text": "#e0e6ed"
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"]
      }
    },
  },
  plugins: [require('daisyui')],
};
