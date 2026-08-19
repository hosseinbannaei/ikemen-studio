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
          900: '#0f1117',
          800: '#161922',
          700: '#1f2430',
          600: '#2b3242',
        }
      }
    },
  },
  plugins: [],
}
