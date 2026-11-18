/** Palette et typographies du design « Shodan@Home — version grand public » (style C) */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        page: '#F3F5FA',
        ink: '#0F1B33',
        body: '#3A4660',
        muted: '#5A6680',
        line: '#C8D0E2',
        hairline: '#E6EAF3',
        tint: '#F0F3FA',
        track: '#E3E8F5',
        brand: { DEFAULT: '#2D5BFF', ink: '#2448D8', dark: '#16308F', soft: '#E6ECFF', line: '#8EA6FF' },
        urgent: { DEFAULT: '#C2410C', soft: '#FFE4D6', tint: '#FFEDE3', ink: '#9A3412', line: '#E9A27F' },
        check: { DEFAULT: '#D89A1E', soft: '#FFF3D6', ink: '#7A4A00', line: '#EBCB85' },
        fine: { soft: '#E4F4EC', ink: '#14532D' },
        off: { DEFAULT: '#A9B2C6', soft: '#EEF0F4' },
      },
      fontFamily: {
        display: ['"Bricolage Grotesque Variable"', 'system-ui', 'sans-serif'],
        sans: ['"Inter Variable"', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
