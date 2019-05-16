<p align="center"><a href="https://orbit.sh"><img src="docs/design/assets/logo/gradient-text-vertical.svg" width="100px" alt="Orbit Logo"></a></p>

<p align="center"><i>A simple and scalable self-hosted Platform as a Service</i></p>

<p align="center">
	<img src="https://img.shields.io/badge/build-passing-brightgreen.svg">
	<img src="https://img.shields.io/badge/coverage-0%25-yellow.svg">
	<img src="https://img.shields.io/badge/version-alpha-orange.svg">
	<a href="https://github.com/prettier/prettier"><img src="https://img.shields.io/badge/code_style-prettier-ff69b4.svg"></a>
	<a href="https://choosealicense.com/licenses/mit/"><img src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</p>

<p align="center"><code>curl https://get.orbit.sh | bash</code></p>

<p align="center">Orbit is designed to take the pain away from self-hosted web applications. It leverages many existing open source container clustering and distributed data solutions to be able to provide a <i>5 minute setup</i> with minimal interaction with the command line. In fact, you only have to run that <i>one</i> command shown above!</p>

## Key features

1. **Five minute set up**: Orbit only requires a single `curl | bash` command to install on your linux server, and the rest of the short set up takes place on secure web dashboard.

2. **Easy to understand**: The core system leverages many existing robust open-source technologies and abstracts away the pain of configuring them manually; making your application much easier to reason about.

3. **Web dashboard**: Included is a _very_ slick, powerful, and secure web dashboard that can do _everything_ without you ever having to touch the command-line if you don't want to.

4. **HTTP(S) edge-routing**: With built-in [LetsEncrypt](https://letsencrypt.org/) for obtaining SSL certificates, Orbit can use host-based routing to direct network traffic to your applications simply and quickly (along with automatic `www.` redirection too).

5. **Automatic deployment**: Borrowing the [buildpacks](https://www.heroku.com/elements/buildpacks) from [Heroku](https://heroku.com/), you can deploy server-side code (using [`git push`](https://devcenter.heroku.com/articles/git#deploying-code), [GitHub](https://developer.github.com/webhooks/), or [GitLab](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html) webhooks) written in any programming language with an automatic detection and build process.

6. **Data included**: Building on top of [GlusterFS](https://docs.gluster.org/en/latest/), out of the box orbit provides highly configurable, distributed, and fault tolerant block storage volumes and databases (such as [Postgres](https://www.postgresql.org/), [Redis](https://redis.io/) & [MySQL](https://www.mysql.com/)).

7. **Omnidirectional scalability**: There is no upper or lower limit to the number of nodes that you can reliably use. For a simple, low-stakes project, one node works fine. For a large, multi-tenant operation, you have the freedom to scale up to thousands of nodes.

## Installation

1. Grab a fresh Ubuntu `18.04` server with a public IP address (ideally a cloud VPS from [Linode](https://www.linode.com/) or [DigitalOcean](https://www.digitalocean.com/)). It can be the smallest offering from either; Orbit is pretty resource light.
2. Log in to the server via ssh as the `root` user.
3. Run `curl https://get.orbit.sh | bash` and wait for about 10 minutes for the installation to complete.
4. Navigate to the IP address of your server at port `6500` (for example, if your server IP address is `121.31.232.56`, you would access the setup page at `http://121.31.232.56:6500`).
