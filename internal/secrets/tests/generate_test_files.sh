#!/bin/bash

set -ueo pipefail

# This file generates some RSA keys that are used for unit tests.  
#
# NONE OF THESE TEST KEYS SHOULD BE USED IN A PRODUCTION SYSTEM.
# YOU SHOULD GENERATE YOUR OWN THAT MEET YOUR SECURITY REQURIEMENTS.
# PRIVATE KEYS IN PRODUCTIONS SYSTEMS SHOULD BE ADEQUATELY PROTECTED WITH A PASSPHRASE

#make a private key encrypted with pkcs#5 v2 ciphers
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:4096 \
    | openssl pkcs8 -topk8 -v2 aes-256-cbc -v2prf hmacWithSHA256 \
              -out key-4096-with-pw.pem -passout pass:testpassword!
openssl rsa -in key-4096-with-pw.pem -passin pass:testpassword! -pubout > key-4096-with-pw.pub

#make a bunch of other unencrypted test keys
openssl genpkey -out key-4096.pem -algorithm RSA -pkeyopt rsa_keygen_bits:4096
openssl rsa -in key-4096.pem -pubout > key-4096.pub

openssl genpkey -out key-2048.pem -algorithm RSA -pkeyopt rsa_keygen_bits:2048
openssl rsa -in key-2048.pem -pubout > key-2048.pub

openssl genpkey -out key-512.pem -algorithm RSA -pkeyopt rsa_keygen_bits:512
openssl rsa -in key-512.pem -pubout > key-512.pub

# generate keys with unsupported cipher
openssl ecparam -name secp256k1 -genkey -noout -out key-EC.pem
openssl ec -in key-EC.pem -pubout > key-EC.pub

echo "-----BEGIN PRIVATE KEY-----
not a valid key file
-----END PRIVATE KEY-----" > not-a-key.txt

echo "All test keys were generated successfully"