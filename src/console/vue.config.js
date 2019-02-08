module.exports = {
  devServer: {
    port: 6500,
    proxy: {
      "/api": {
        target: { socketPath: "/var/run/orbit.sock" },
        changeOrigin: false,
        pathRewrite: { "/api": "" }
      }
    }
  }
};
