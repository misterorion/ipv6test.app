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
        'bg-pink',
        'border-pink',
        'bg-purple',
        'border-purple',
        'bg-yellow',
        'border-yellow',
        'text-black',
        'text-white'
      ]
    }
  }
}