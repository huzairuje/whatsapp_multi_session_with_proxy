/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#25D366',
        secondary: '#128C7E',
        success: '#34D399',
        warning: '#FBBF24',
        error: '#EF4444',
      },
    },
  },
  plugins: [],
}
