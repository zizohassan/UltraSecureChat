# Generate server private key
openssl genpkey -algorithm RSA -out server-key.pem -pkeyopt rsa_keygen_bits:4096

# Generate server certificate signing request (CSR)
openssl req -new -key server-key.pem -out server.csr -config server_openssl.cnf

# Sign the server certificate with the CA
openssl x509 -req -in server.csr -CA cert/root-cert.pem -CAkey cert/root-key.pem -CAcreateserial -out server-cert.pem -days 365 -extensions v3_req -extfile server_openssl.cnf

mv server-key.pem cert/server-key.pem
mv server-cert.pem cert/server-cert.pem
mv server.csr cert/server.csr