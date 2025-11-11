// tailwind.config.cjs
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}",
    "./web/preview/**/*.html" // if you keep preview html under web/preview
  ],
  theme: {
    extend: {
      colors: {
        voidbg: "#0b0b11",
        accent: "#6cf",
        danger: "#ff6b6b",
      },
      fontFamily: {
        mono2: ["'Courier New'", "monospace"]
      }
    }
  },
  plugins: []
};
