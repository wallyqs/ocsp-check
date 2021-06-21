# OCSP Response Check

## Quick install

```
curl -sf https://gobinaries.com/wallyqs/ocsp-check | PREFIX=. sh

./ocsp-check -h
Usage: nats-ocsp-resp [-s server] [-creds file] [-nkey file] [-tlscert file] [-tlskey file] [-tlscacert file]
  -creds string
    	User Credentials File
  -h	Show help message
  -nkey string
    	NKey Seed File
  -s string
    	The nats server URLs (separated by comma) (default "nats://127.0.0.1:4222")
  -tlscacert string
    	CA certificate to verify peer against
  -tlscert string
    	TLS client certificate file
  -tlskey string
    	Private key file for client certificate
```

## Example usage

```
ocsp-check -s localhost:4222 -tlscacert /tmp/ca-cert.pem 
```

Result: 

```
--- NATS OCSP Response ---
Status: Good
ProducedAt:  2021-06-21 21:59:00 +0000 UTC
ThisUpdate:  2021-06-21 21:59:29 +0000 UTC
NextUpdate:  2021-06-21 21:59:33 +0000 UTC
```
