module.exports = {
  purge: {
    enabled: true,
    content: [
      "./app/views/**/*.html"
    ]
  },
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {},
  },
  variants: {
    extend: {},
  },
  plugins: [
    require("@tailwindcss/forms")
  ]
}
