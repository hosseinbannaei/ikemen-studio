/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          900: 'var(--bg-app)',
          850: 'var(--bg-sidebar)',
          800: 'var(--bg-surface)',
          750: 'var(--bg-hover)',
          700: 'var(--bg-card)',
          600: 'var(--bg-border)',
        }
      }
    },
  },
  plugins: [],
}
