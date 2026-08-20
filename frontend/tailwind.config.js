/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      borderColor: {
        DEFAULT: 'rgb(var(--bg-border-rgb) / <alpha-value>)',
      },
      colors: {
        dark: {
          900: 'rgb(var(--bg-app-rgb) / <alpha-value>)',
          850: 'rgb(var(--bg-sidebar-rgb) / <alpha-value>)',
          800: 'rgb(var(--bg-surface-rgb) / <alpha-value>)',
          750: 'rgb(var(--bg-hover-rgb) / <alpha-value>)',
          700: 'rgb(var(--bg-card-rgb) / <alpha-value>)',
          600: 'rgb(var(--bg-border-rgb) / <alpha-value>)',
        }
      }
    },
  },
  plugins: [],
}
