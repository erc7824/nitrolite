/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  prefix: "",
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
      },
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // Brand colors
        brand: {
          cyan: {
            DEFAULT: "#00e5ff",
            dark: "#0099cc",
            light: "#80f2ff",
          },
          magenta: {
            DEFAULT: "#ff49e1",
            dark: "#cc00aa",
            light: "#ffb2f0",
          },
          navy: {
            DEFAULT: "#070720",
            light: "#141432",
            lighter: "#1e1e46",
          },
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      boxShadow: {
        glow: {
          cyan: "0 0 20px 5px rgba(0, 229, 255, 0.3)",
          magenta: "0 0 20px 5px rgba(255, 73, 225, 0.3)",
        },
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
        pulse: {
          "0%, 100%": { opacity: 0.8 },
          "50%": { opacity: 1 },
        },
        sparkle: {
          "0%": { transform: "translateY(0) rotate(0deg)" },
          "100%": { transform: "translateY(-100px) rotate(20deg)" },
        },
        float: {
          "0%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-10px)" },
          "100%": { transform: "translateY(0px)" },
        },
        "pulse-cyan": {
          "0%": { boxShadow: "0 0 10px rgba(0, 229, 255, 0.5)" },
          "50%": { 
            boxShadow: "0 0 20px rgba(0, 229, 255, 0.7), 0 0 30px rgba(0, 229, 255, 0.3)" 
          },
          "100%": { boxShadow: "0 0 10px rgba(0, 229, 255, 0.5)" },
        },
        "pulse-magenta": {
          "0%": { boxShadow: "0 0 10px rgba(255, 73, 225, 0.5)" },
          "50%": { 
            boxShadow: "0 0 20px rgba(255, 73, 225, 0.7), 0 0 30px rgba(255, 73, 225, 0.3)" 
          },
          "100%": { boxShadow: "0 0 10px rgba(255, 73, 225, 0.5)" },
        },
        "explode-cyan": {
          "0%": { 
            transform: "scale(0.3)",
            opacity: 1,
            filter: "blur(2px)"
          },
          "80%": {
            opacity: 0.7
          },
          "100%": { 
            transform: "scale(3)", 
            opacity: 0,
            filter: "blur(1px)"
          },
        },
        "explode-magenta": {
          "0%": { 
            transform: "scale(0.3)",
            opacity: 1,
            filter: "blur(2px)"
          },
          "80%": {
            opacity: 0.7
          },
          "100%": { 
            transform: "scale(3)", 
            opacity: 0,
            filter: "blur(1px)"
          },
        },
        "orbit": {
          "0%": { transform: "rotate(0deg) translateX(10px) rotate(0deg)" },
          "100%": { transform: "rotate(360deg) translateX(10px) rotate(-360deg)" },
        },
        fadeIn: {
          "0%": { opacity: 0 },
          "100%": { opacity: 1 },
        },
        fadeInUp: {
          "0%": { 
            opacity: 0,
            transform: "translateY(10px)",
          },
          "100%": { 
            opacity: 1,
            transform: "translateY(0)",
          },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "pulse": "pulse 3s infinite ease-in-out",
        "pulse-fast": "pulse 1.5s infinite ease-in-out",
        "sparkle": "sparkle 8s infinite linear",
        "float": "float 6s infinite ease-in-out",
        "pulse-cyan": "pulse-cyan 2s infinite ease-in-out",
        "pulse-magenta": "pulse-magenta 2s infinite ease-in-out",
        "explode-cyan": "explode-cyan 1.5s ease-out forwards",
        "explode-magenta": "explode-magenta 1.5s ease-out forwards",
        "orbit": "orbit 3s infinite linear",
        "orbit-reverse": "orbit 3s infinite linear reverse",
        "fadeIn": "fadeIn 0.3s ease-in-out",
        "fadeInUp": "fadeInUp 0.5s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}