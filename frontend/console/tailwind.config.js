/** @type {import('tailwindcss').Config} */
import forms from '@tailwindcss/forms'

export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}', './.storybook/**/*.{js,jsx,ts,tsx}'],
  darkMode: ['class'],
  theme: {
    extend: {
      fontFamily: {
        'roboto-mono': ['Roboto Mono', 'monospace'],
      },
      keyframes: {
        appear: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'rain-1': {
          '0%': { opacity: '0', transform: 'translateY(0)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0', transform: 'translateY(8px)' },
        },
        'rain-2': {
          '0%': { opacity: '0', transform: 'translateY(-4px)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0', transform: 'translateY(4px)' },
        },
        'rain-3': {
          '0%': { opacity: '0', transform: 'translateY(-8px)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0', transform: 'translateY(0)' },
        },
        'rain-4': {
          '0%': { opacity: '0', transform: 'translateY(4px)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0', transform: 'translateY(12px)' },
        },
        'rain-5': {
          '0%': { opacity: '0', transform: 'translateY(-2px)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0', transform: 'translateY(6px)' },
        },
      },
      animation: {
        appear: 'appear 250ms ease-in forwards',
        'rain-1': 'rain-1 1.5s ease-in-out infinite',
        'rain-2': 'rain-2 1.5s ease-in-out infinite 0.2s',
        'rain-3': 'rain-3 1.5s ease-in-out infinite 0.4s',
        'rain-4': 'rain-4 1.5s ease-in-out infinite 0.6s',
        'rain-5': 'rain-5 1.5s ease-in-out infinite 0.8s',
      },
    },
  },
  plugins: [forms],
}
