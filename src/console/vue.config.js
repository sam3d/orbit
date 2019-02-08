module.exports = {
  devServer: {
    port: 6500,
    watchOptions: { poll: true },
    proxy: {
      "/api": {
        target: { socketPath: "/var/run/orbit.sock" },
        changeOrigin: false,
        pathRewrite: { "/api": "" }
      }
    }
  }
};
