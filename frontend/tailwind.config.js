/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      borderRadius: {
        'none': '0px',
        'sm': 'var(--radius-sm)',
        'DEFAULT': 'var(--radius-md)',
        'md': 'var(--radius-md)',
        'lg': 'var(--radius-lg)',
        'xl': 'var(--radius-xl)',
        '2xl': 'var(--radius-2xl)',
        '3xl': 'var(--radius-3xl)',
        'full': '9999px',
      },
      borderColor: {
        DEFAULT: 'rgb(var(--bg-border-rgb) / <alpha-value>)',
      },
      colors: {
        dark: {
          950: 'rgb(var(--bg-app-darker-rgb) / <alpha-value>)',
          900: 'rgb(var(--bg-app-rgb) / <alpha-value>)',
          850: 'rgb(var(--bg-sidebar-rgb) / <alpha-value>)',
          800: 'rgb(var(--bg-surface-rgb) / <alpha-value>)',
          750: 'rgb(var(--bg-hover-rgb) / <alpha-value>)',
          700: 'rgb(var(--bg-card-rgb) / <alpha-value>)',
          600: 'rgb(var(--bg-border-rgb) / <alpha-value>)',
        },
        brand: {
          400: 'rgb(var(--brand-400-rgb) / <alpha-value>)',
          500: 'rgb(var(--brand-500-rgb) / <alpha-value>)',
          600: 'rgb(var(--brand-600-rgb) / <alpha-value>)',
          700: 'rgb(var(--brand-700-rgb) / <alpha-value>)',
          950: 'rgb(var(--brand-950-rgb) / <alpha-value>)',
        }
      }
    },
  },
  plugins: [],
}
