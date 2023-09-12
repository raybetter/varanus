# varanus
monitoring tool



## Make a key set

```sh
openssl genpkey -out private-key.pem -algorithm RSA -pkeyopt rsa_keygen_bits:4096
openssl rsa -in private-key.pem -pubout > public-key.pub
```