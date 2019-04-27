// Package nginx implements basic nginx configuration generation primitives for
// the purpose of loading the data correctly on start up.
package nginx

import (
	"fmt"
	"strings"
)

// App is a logical group of multiple server blocks to serve the purpose of what
// needs to occur in the case of routing.
type App struct {
	Domain  string
	ProxyTo string

	HTTPS              bool
	CertificateFile    string
	CertificateKeyFile string

	WWWRedirect bool
}

// Marshal generates a complete set of server blocks for the specified app
// configuration.
func (a App) Marshal() string {
	b := ""

	if a.HTTPS {
		b += a.httpsRedirect() + "\n\n"
	}

	if a.WWWRedirect {
		b += a.wwwRedirect() + "\n\n"
	}

	b += a.proxyPass()

	return b
}

// httpsRedirect will create a block that redirects to HTTPS.
func (a App) httpsRedirect() string {
	return fmt.Sprintf(`server {
  listen 80;
  listen [::]:80;
  server_name %s;

  return 301 https://$host$request_uri;
}`, a.Domain)
}

/// wwwRedirect will create a block that redirects to the www or non-www version
//of a domain name.
func (a App) wwwRedirect() string {
	// If the main domain starts with www, we want to redirect from the non "www"
	// version to the one that has it. However, if the main domain does not start
	// with "www", we want to redirect from the "www" version to it.
	var src string
	if strings.HasPrefix(a.Domain, "www.") {
		src = strings.TrimPrefix(a.Domain, "www.")
	} else {
		src = "www." + a.Domain
	}

	// Start the creation of the server block.
	b := "server {\n"

	// Add HTTP listener.
	b += "  listen 80;\n  listen [::]:80;\n\n"

	// If using HTTPS, add the HTTPS listener to redirect every request to the
	// correct URL.
	if a.HTTPS {
		b += "  listen 443 ssl;\n  listen [::]:443 ssl;\n\n"
	}

	// Add the server name.
	b += "  server_name " + src + ";\n\n"

	// Add the SSL certificate details if using HTTPS.
	if a.HTTPS {
		b += "  ssl_certificate " + a.CertificateFile + ";\n"
		b += "  ssl_certificate_key " + a.CertificateKeyFile + ";\n\n"
	}

	// Add the redirect handler.
	var protocol string
	if a.HTTPS {
		protocol = "https://"
	} else {
		protocol = "http://"
	}
	b += fmt.Sprintf("  return 301 %s%s$request_uri;\n", protocol, a.Domain)

	// Close the server block and return.
	b += "}"
	return b
}

func (a App) proxyPass() string {
	// Start the server block.
	b := "server {\n"

	// Add the correct listener.
	if !a.HTTPS {
		b += "  listen 80;\n  listen [::]:80;\n"
	} else {
		b += "  listen 443 ssl;\n  listen [::]:443 ssl;\n"
	}

	// Add the server name.
	b += "  server_name " + a.Domain + ";\n\n"

	// Add the SSL certificate details if using HTTPS.
	if a.HTTPS {
		b += "  ssl_certificate " + a.CertificateFile + ";\n"
		b += "  ssl_certificate_key " + a.CertificateKeyFile + ";\n"
	}

	// Add the location block.
	b += fmt.Sprintf(`
  location / {
    proxy_pass http://%s;

    proxy_redirect     off;
    proxy_set_header   Host $host;
    proxy_set_header   X-Real-IP $remote_addr;
    proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header   X-Forwarded-Host $server_name;
  }
`, a.ProxyTo)

	// Close the server block and return.
	b += "}"
	return b
}
