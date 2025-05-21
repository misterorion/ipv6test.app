module.exports = {
  plugins: {
    '@tailwindcss/postcss': {},
    cssnano: {
      preset: ["default", { discardComments: { removeAll: true } }],
    }
  }
}