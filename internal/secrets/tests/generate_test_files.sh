#!/bin/bash

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