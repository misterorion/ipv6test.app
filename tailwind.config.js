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
        'border-pink',
        'bg-pink',
        'border-blue',
        'bg-blue'
      ]
    }
  },
}