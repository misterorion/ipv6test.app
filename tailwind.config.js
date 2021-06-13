module.exports = {
  plugins: [],
  purge: {
    enabled: true,
    mode: 'all',
    content: [
      './srv/index.html',
    ],
    options: {
      safelist: [
        'border-green-700',
        'bg-green-700',
        'border-red-700',
        'bg-red-700'
      ]
    }
  },
}