/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './priv/static/**/*.html',
    './lib/personal_blog_web/**/*.*ex',
  ],
  theme: {
    extend: {
      typography: {
        DEFAULT: {
          css: {
            maxWidth: '100%',
          }
        }
      }
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
