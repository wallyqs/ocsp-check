// Copyright 2021 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
	"golang.org/x/crypto/ocsp"
)

func init() {
	log.SetFlags(0)
}

func usage() {
	log.Printf("Usage: nats-ocsp-resp [-s server] [-creds file] [-nkey file] [-tlscert file] [-tlskey file] [-tlscacert file] <subject> <msg>\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func getOCSPStatus(s tls.ConnectionState) (*ocsp.Response, error) {
	if len(s.VerifiedChains) == 0 {
		return nil, fmt.Errorf("missing TLS verified chains")
	}
	chain := s.VerifiedChains[0]

	if got, want := len(chain), 2; got < want {
		return nil, fmt.Errorf("incomplete cert chain, got %d, want at least %d", got, want)
	}
	leaf, issuer := chain[0], chain[1]

	resp, err := ocsp.ParseResponseForCert(s.OCSPResponse, leaf, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OCSP response: %w", err)
	}
	if err := resp.CheckSignatureFrom(issuer); err != nil {
		return resp, fmt.Errorf("bad OCSP signature: %v", err)
	}
	return resp, nil
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var userCreds = flag.String("creds", "", "User Credentials File")
	var nkeyFile = flag.String("nkey", "", "NKey Seed File")
	var tlsClientCert = flag.String("tlscert", "", "TLS client certificate file")
	var tlsClientKey = flag.String("tlskey", "", "Private key file for client certificate")
	var tlsCACert = flag.String("tlscacert", "", "CA certificate to verify peer against")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS OCSP Response Check")}

	if *userCreds != "" && *nkeyFile != "" {
		log.Fatal("specify -seed or -creds")
	}

	// Use UserCredentials
	if *userCreds != "" {
		opts = append(opts, nats.UserCredentials(*userCreds))
	}

	// Use TLS client authentication
	if *tlsClientCert != "" && *tlsClientKey != "" {
		opts = append(opts, nats.ClientCert(*tlsClientCert, *tlsClientKey))
	}

	// Check OCSP Response
	secure := nats.Secure(&tls.Config{
		VerifyConnection: func(s tls.ConnectionState) error {
			resp, err := getOCSPStatus(s)
			if resp != nil {
				log.Println("--- NATS OCSP Response ---")
				switch resp.Status {
				case ocsp.Good:
					log.Println("Status: Good")
				case ocsp.Revoked:
					log.Println("Status: Revoked")
					log.Println("RevokedAt : ", resp.RevokedAt)
				case ocsp.Unknown:
					log.Println("Status: Unknown")
				default:
					return fmt.Errorf("invalid staple status")
				}
				if resp.Status != ocsp.Good {
					return fmt.Errorf("invalid staple")
				}
				log.Println("ProducedAt: ", resp.ProducedAt)
				log.Println("ThisUpdate: ", resp.ThisUpdate)
				log.Println("NextUpdate: ", resp.NextUpdate)
				log.Println("")
				log.Println("--- OCSP Signature ---")
				log.Printf("Verified: %v", err == nil)
				log.Println("")
				log.Println("--- Signature Algorithms ---")
				log.Println("OCSP Response     :", resp.SignatureAlgorithm)

				chain := s.VerifiedChains[0]
				leaf, issuer := chain[0], chain[1]
				log.Printf("Leaf Certificate  : %+v", leaf.SignatureAlgorithm)
				log.Printf("Issuer Certificate: %+v", issuer.SignatureAlgorithm)
			}
			if err != nil {
				return err
			}
			return nil
		},
	})
	opts = append(opts, secure)

	// Use specific CA certificate
	if *tlsCACert != "" {
		opts = append(opts, nats.RootCAs(*tlsCACert))
	}

	// Use Nkey authentication.
	if *nkeyFile != "" {
		opt, err := nats.NkeyOptionFromSeed(*nkeyFile)
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, opt)
	}

	// Connect to NATS
	nc, err := nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer nc.Close()
}
