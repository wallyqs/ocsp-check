# OCSP Response Check

Example usage:

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
