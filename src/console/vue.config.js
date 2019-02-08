module.exports = {
  devServer: {
    port: 3000,
    proxy: {
      "/api": {
        target: { socketPath: "/var/run/orbit.sock" },
        changeOrigin: false,
        pathRewrite: { "/api": "" }
      }
    }
  }
};
