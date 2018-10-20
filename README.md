# fibr

Web File Browser and Manager

[![Build Status](https://travis-ci.org/ViBiOh/fibr.svg?branch=master)](https://travis-ci.org/ViBiOh/fibr)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/fibr)](https://goreportcard.com/report/github.com/ViBiOh/fibr)

Thanks to [FontAwesome](https://fontawesome.com) for providing awesome svg.

## Installation

```bash
go get github.com/ViBiOh/fibr/cmd
```

## Usage

```bash
Usage of fibr:
  -authDisable
      [auth] Disable auth
  -authUrl string
      [auth] Auth URL, if remote
  -authUsers string
      [auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)
  -basicUsers string
      [Basic] Users in the form "id:username:password,id2:username2:password2"
  -csp string
      [owasp] Content-Security-Policy (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -fsDirectory string
      [filesystem] Path to served directory (default "/data")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -metadata
      Enable metadata storage (default true)
  -minioAccessKey string
      [minio] Access Key
  -minioEndpoint string
      [minio] Endpoint server
  -minioSecretKey string
      [minio] Secret Key
  -port int
      Listen port (default 1080)
  -prometheusPath string
      [prometheus] Path for exposing metrics (default "/metrics")
  -publicURL string
      [fibr] Public URL (default "https://fibr.vibioh.fr")
  -storage string
      Storage used (e.g. 'filesystem', 'minio') (default "filesystem")
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -tlsOrganization string
      [tls] Self-signed certificate organization (default "ViBiOh")
  -tracingAgent string
      [opentracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
      [opentracing] Service name
  -url string
      [health] URL to check
  -userAgent string
      [health] User-Agent for check (default "Golang alcotest")
  -version string
      [fibr] Version (used mainly as a cache-buster)
```
