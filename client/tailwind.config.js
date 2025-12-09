/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
        ornate: "0.5rem",
      },
      colors: {
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        chart: {
          "1": "hsl(var(--chart-1))",
          "2": "hsl(var(--chart-2))",
          "3": "hsl(var(--chart-3))",
          "4": "hsl(var(--chart-4))",
          "5": "hsl(var(--chart-5))",
        },
        // Casino-themed custom colors
        felt: {
          DEFAULT: "hsl(var(--felt))",
          dark: "hsl(var(--felt-dark))",
        },
        gold: {
          DEFAULT: "hsl(var(--primary))",
          light: "hsl(var(--gold-light))",
          dark: "hsl(43 55% 45%)",
        },
        mahogany: {
          DEFAULT: "hsl(var(--background))",
          light: "hsl(var(--card))",
          dark: "hsl(20 50% 10%)",
        },
      },
      fontFamily: {
        display: ['Cinzel', 'serif'],
        heading: ['Playfair Display', 'serif'],
        elegant: ['Cormorant Garamond', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        'gold': '0 0 20px rgba(212, 168, 85, 0.3), 0 0 40px rgba(212, 168, 85, 0.15)',
        'gold-lg': '0 0 30px rgba(212, 168, 85, 0.4), 0 0 60px rgba(212, 168, 85, 0.2)',
        'inset-deep': 'inset 0 2px 8px rgba(0, 0, 0, 0.4)',
        'raised': '0 1px 0 rgba(255, 255, 255, 0.1), 0 2px 4px rgba(0, 0, 0, 0.3)',
        'poker-chip': '0 4px 8px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.2)',
      },
      animation: {
        'glow-pulse': 'glow-pulse 2s ease-in-out infinite',
        'lift': 'lift 0.3s ease-out',
      },
      keyframes: {
        'glow-pulse': {
          '0%, 100%': {
            boxShadow: '0 0 20px rgba(212, 168, 85, 0.3)',
          },
          '50%': {
            boxShadow: '0 0 40px rgba(212, 168, 85, 0.5)',
          },
        },
        'lift': {
          '0%': { transform: 'translateY(0)' },
          '100%': { transform: 'translateY(-4px)' },
        },
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}
