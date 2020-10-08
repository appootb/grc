module.exports = {
  "transpileDependencies": [
    "vuetify"
  ],
  devServer: {
    proxy: {
      "/api": {
        target: "http://localhost:12345",
        changeOrigin: true,
      }
    }
  }
}