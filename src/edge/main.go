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

	// Write the certificates.
	os.MkdirAll("/etc/nginx/certs", os.ModePerm)
	for _, c := range certificates {
		certificatePath := filepath.Join(CertsPath, c.ID+".crt")
		privateKeyPath := filepath.Join(CertsPath, c.ID+".key")

		if err := ioutil.WriteFile(certificatePath, c.FullChain, 0644); err != nil {
			log.Fatalf("Could not write certificate: %s", err)
		}
		if err := ioutil.WriteFile(privateKeyPath, c.PrivateKey, 0644); err != nil {
			log.Fatalf("Could not write private key: %s", err)
		}
	}

	// Generate nginx app objects.
	var apps []nginx.App
	for _, r := range routers {
		// Create the standard app.
		app := nginx.App{
			Domain:  r.Domain,
			ProxyTo: r.AppID,
		}

		// If it uses HTTPS, add the certificate details.
		if r.CertificateID != "" {
			app.HTTPS = true
			app.CertificateFile = filepath.Join(CertsPath, r.CertificateID+".crt")
			app.CertificateKeyFile = filepath.Join(CertsPath, r.CertificateID+".key")
		}

		// TODO: Set this to use an actual router value.
		app.WWWRedirect = true

		// Actually add the app to the apps list.
		apps = append(apps, app)
	}

	// Generate the config.
	var config string
	for _, a := range apps {
		config += a.Marshal() + "\n\n"
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
