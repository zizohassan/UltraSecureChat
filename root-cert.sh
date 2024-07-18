# Generate CA private key
openssl genpkey -algorithm RSA -out root-key.pem -pkeyopt rsa_keygen_bits:4096

# Generate CA certificate
openssl req -x509 -new -key root-key.pem -out root-cert.pem -days 365 -subj "/CN=My Root CA"

cp root-cert.pem build/Chat\ Client.app/Contents/Resources/root-cert.pem
cp root-cert.pem build/root-cert.pem
mv root-key.pem cert/root-key.pem
mv root-cert.pem cert/root-cert.pem