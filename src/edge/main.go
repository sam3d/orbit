package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"orbit.sh/edge/nginx"
)

const (
	// CertsPath is the directory where the SSL certificates are kept.
	CertsPath = "/etc/nginx/certs"

	// ChallengeFile is the file where the challenges (as "location" directives)
	// are kept. This must be included only in "server" blocks, as location blocks
	// cannot exist outside of that context.
	ChallengeFile = "challenges.conf"
)

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

	// Now write the challenges into their own file to be included in every app.
	var challengeConfig string
	for _, c := range challenges {
		challengeConfig += nginx.GenerateLocation(c.Path, c.Token)
	}
	if err := ioutil.WriteFile(filepath.Join(CertsPath, ChallengeFile), []byte(challengeConfig), 0644); err != nil {
		log.Fatalf("Could not write challenge file: %s", err)
	}

	// Prepare the config string. This will be the final configuration string that
	// gets written to the configuration file, so any and all nginx configuration
	// needs to go in here.
	var config string

	// Add the default 404 catch-all handler.
	config += nginx.GenerateDefault() + "\n\n"

	// Loop over all of the router objects and create their properties.
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

		// Actually add the app to the config.
		config += app.Marshal() + "\n\n"
	}

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
