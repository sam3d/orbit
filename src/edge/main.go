package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"orbit.sh/edge/nginx"
)

// CertsPath is the directory where the SSL certificates are kept.
const CertsPath = "/etc/nginx/certs"

func main() {
	client := NewClient()

	routers := client.GetRouters()
	certificates := client.GetCertificates()
	var challenges []Challenge

	// Write the certificates and the challenges.
	os.MkdirAll("/etc/nginx/certs", os.ModePerm)
	for _, c := range certificates {
		// Write the certificate and the private key.
		certificatePath := filepath.Join(CertsPath, c.ID+".crt")
		privateKeyPath := filepath.Join(CertsPath, c.ID+".key")

		if err := ioutil.WriteFile(certificatePath, c.FullChain, 0644); err != nil {
			log.Fatalf("Could not write certificate: %s", err)
		}
		if err := ioutil.WriteFile(privateKeyPath, c.PrivateKey, 0644); err != nil {
			log.Fatalf("Could not write private key: %s", err)
		}

		// Add the challenges to the challenges variable.
		challenges = append(challenges, c.Challenges...)
	}

	// Generate nginx app objects.
	var apps []nginx.App
	for _, r := range routers {
		// Create the standard app.
		app := nginx.App{
			Domain:      r.Domain,
			ProxyTo:     r.AppID,
			WWWRedirect: r.WWWRedirect,
		}

		// If it uses HTTPS, add the certificate details. We need to perform the
		// check to ensure that we have every bit of detail required before we can
		// go adding HTTPS. This is primarily to ensure that we don't try to enable
		// HTTPS and then have nginx throw a fit because it can't find or verify the
		// SSL certificates.
		if r.CertificateID != "" && app.CertificateFile != "" && app.CertificateKeyFile != "" {
			app.HTTPS = true
			app.CertificateFile = filepath.Join(CertsPath, r.CertificateID+".crt")
			app.CertificateKeyFile = filepath.Join(CertsPath, r.CertificateID+".key")
		}

		// Actually add the app to the apps list.
		apps = append(apps, app)
	}

	// Generate the config from the apps.
	var config string
	for _, a := range apps {
		config += a.Marshal() + "\n\n"
	}

	// Append the challenges to the config. This also includes the default server
	// for handling any 404 requests. They must go in the same server block so
	// that it serves them as a catch-all route handler.
	config += `server {
  listen 80 default_server;
  listen [::]:80 default_server;

  listen 443 default_server ssl;
  listen [::]:443 default_server ssl;
  ssl_certificate /etc/nginx/certs/dummy/cert.pem;
  ssl_certificate_key /etc/nginx/certs/dummy/key.pem;

  server_name _;

  location / {
    return 404;
  }

`
	for _, c := range challenges {
		config += fmt.Sprintf("  location %s { add_header Content-Type text/plain; return 200 \"%s\"; }\n", c.Path, c.Token)
	}

	config += "}"

	// Write the config.
	os.MkdirAll("/etc/nginx/conf.d", os.ModePerm)
	if err := ioutil.WriteFile("/etc/nginx/conf.d/orbit.conf", []byte(config), 0644); err != nil {
		log.Fatalf("Could not write the config file: %s", err)
	}
}

func example() {
	a := nginx.App{
		Domain:             "orbit.samholmes.net",
		ProxyTo:            "app",
		HTTPS:              true,
		WWWRedirect:        true,
		CertificateFile:    "/etc/test",
		CertificateKeyFile: "/etc/test.key",
	}

	fmt.Println(a.Marshal())
}
