package main

import (
	"fmt"

	"orbit.sh/edge/nginx"
)

func main() {
	example()
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
