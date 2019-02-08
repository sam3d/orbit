const logger = require("@vue/cli-shared-utils/lib/logger");

/**
 * This will set the API URL for Orbit. By default, it will use the environment
 * variable "ORBIT_API_URL". If that is not set, it will check whether not this
 * is a production environment and use variable accordingly.
 *
 * NOTE: In a production environment, the HTTP socket is bound via a volume
 * mount. If in a development environment, socket mounts can be hard for
 * individual platforms or environments, so by default will use the exposed HTTP
 * port that Orbit uses for accessing the API. Both are enabled by default, but
 * using a UNIX socket is easier to bind than using the HTTP when it comes to
 * Docker.
 */
const API_URL = process.env.ORBIT_API_URL
  ? process.env.ORBIT_API_URL
  : process.env.NODE_ENV === "production"
  ? "/api"
  : "http://localhost:6501";
process.env.VUE_APP_ORBIT_API_URL = API_URL;
logger.info(`Using Orbit API URL: "${API_URL}"`);

module.exports = {
  devServer: {
    port: 3000
  }
};
