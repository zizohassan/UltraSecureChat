# Generate client private key
openssl genpkey -algorithm RSA -out client-key.pem -pkeyopt rsa_keygen_bits:4096

# Generate client certificate signing request (CSR)
openssl req -new -key client-key.pem -out client.csr -config client_openssl.cnf

# Sign the client certificate with the CA
openssl x509 -req -in client.csr -CA cert/root-cert.pem -CAkey cert/root-key.pem -CAcreateserial -out client-cert.pem -days 365 -extensions v3_req -extfile client_openssl.cnf


mv client-key.pem cert/client-key.pem
mv client-cert.pem cert/client-cert.pem
mv client.csr cert/client.csr

cp cert/client-key.pem build/client-key.pem
cp cert/client-cert.pem build/client-cert.pem
cp cert/client.csr build/client.csr


cp cert/client-key.pem build/Chat\ Client.app/Contents/Resources/client-key.pem
cp cert/client-cert.pem build/Chat\ Client.app/Contents/Resources/client-cert.pem
cp cert/client.csr build/Chat\ Client.app/Contents/Resources/client.csr