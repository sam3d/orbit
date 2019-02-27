module.exports = {
  devServer: {
    port: 3000,
    watchOptions: { poll: true },
    proxy: {
      "/api": {
        target: process.env.ORBIT_API_URL || "http://localhost:6501",
        changeOrigin: false,
        pathRewrite: { "/api": "" }
      }
    }
  }
};
