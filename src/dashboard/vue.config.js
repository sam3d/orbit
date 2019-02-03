/**
 * Bind the environment variables so that vue cli can interpolate them.
 * https://cli.vuejs.org/guide/mode-and-env.html#using-env-variables-in-client-side-code
 */
["ORBIT_API_URL"].forEach(name => {
  const env = process.env[name];
  if (!env) throw `Missing environment variable: ${name}`;
  process.env["VUE_APP_" + name] = env;
});

module.exports = {
  devServer: { port: 3000 }
};
