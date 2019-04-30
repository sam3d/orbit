module.exports = {
  devServer: {
    port: 3000,
    watchOptions: { poll: true },
    disableHostCheck: true,
    proxy: {
      "/api": {
        target: process.env.ORBIT_API_URL || "http://localhost:6505",
        changeOrigin: false,
        pathRewrite: { "/api": "" }
      }
    }
  }
};
