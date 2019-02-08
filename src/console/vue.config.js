/**
 * Bind the environment variables so that vue cli can interpolate them.
 * https://cli.vuejs.org/guide/mode-and-env.html#using-env-variables-in-client-side-code
 */
const defaultOptions = {
  ORBIT_API_URL: "/api"
};

for (let name in defaultOptions) {
  const env = process.env[name] || defaultOptions[name];
  const key = "VUE_APP_" + name;
  process.env[key] = env;
}

module.exports = {
  devServer: { port: 3000 }
};
