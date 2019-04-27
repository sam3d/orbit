package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"orbit.sh/edge/nginx"
)

func main() {
	client := NewClient()

	routers := client.GetRouters()
	// certificates := client.GetCertificates()

	// Generate nginx app objects.
	var apps []nginx.App
	for _, r := range routers {
		apps = append(apps, nginx.App{
			Domain:      r.Domain,
			ProxyTo:     r.AppID,
			HTTPS:       false,
			WWWRedirect: true,
		})
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
