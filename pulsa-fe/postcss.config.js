// postcss.config.js
// Next.js + Turbopack: Tailwind harus lewat @tailwindcss/postcss

module.exports = {
  plugins: {
    "@tailwindcss/postcss": {},
    autoprefixer: {},
  },
};
