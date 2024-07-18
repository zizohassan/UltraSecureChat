
# UltraSecureChat

## Overview

raSecureChat is a high-security chat application designed to provide end-to-end encryption and secure communication between users. This application leverages AES-256 encryption, secure certificate-based authentication, and a unique combination of dynamic keys and shuffling algorithms to ensure maximum privacy and security. It includes features such as user authentication, encrypted messaging, and the use of unique session keys for each user. Additionally, it provides weather-based dynamic key generation and supports manual session key distribution for enhanced security.


## Key Features

- **Encrypted Messaging**: All messages are encrypted using AES-256, ensuring that only the intended recipients can read the content.
- **Secure Authentication**: The application uses TLS with certificates for client and server authentication, preventing unauthorized access.
- **Dynamic Key Generation**: Unique session keys are generated for each user session, adding an additional layer of security.
- **Weather-Based Key Variation**: Utilizes real-time weather data to generate unique, dynamic keys for encrypting communication, making it highly resistant to pattern recognition attacks.
- **Manual Key Distribution**: Administrators can manually distribute session keys and a 50-digit number for shuffling chat text, ensuring that only trusted users can join the conversation.
- **Secure Coordination**: Generates and validates session keys, and shuffles messages using a unique seed for added security.
- **Randomized Coordinates**: Uses random image coordinates for session initialization, further obfuscating the encryption process.

## Certificate Generation

### Using Local IP

1. update the `CN` and `IP.1` fields in `server_openssl.cnf` with your Local  IP

### Using VPN

1. If using a VPN with port forwarding, update the `CN` and `IP.1` fields in `server_openssl.cnf` with your VPN IP.
2. We recommend using a VPN for an additional security layer.

### Running Certificate Scripts

1. Run `cer.sh` to generate new root, server, and client certificates.
    ```sh
    ./cer.sh
    ```
2. To generate a certificate for each client, use `client-cer.sh`.
    ```sh
    ./client-cer.sh
    ```
3. Share the client certificates with users.

**Note:** If the root certificate changes, you will need to update the server and client certificates accordingly.

## Steps to Connect

### Server Side Steps

1. Place a random image in the `images` directory named `secret.jpeg`.
2. Run `cor.go` to generate coordinates.
    ```sh
    go run cor.go -image="images/secret.jpeg"
    ```
    - Example output:
    ```
    370,1067;1425,149;88,941;90,1166;124,204;1568,128;1236,989;432,511;1477,1164;267,279
    ```
3. Run the server file with the coordinates and image.
    ```sh
    go run server.go -image="images/secret.jpeg" -coords="895,238;604,1158;399,1010;104,181;1358,297;1078,706;854,660;1249,920;1544,366;468,487"
    ```
4. The admin must send session keys manually to each user.
5. The admin must send a 50-digit number to shuffle and decrypt chat text.
6. if it will be local make sure to make port forward on port 443 


### Inviting a User

1. Ask Admin About ip (local ip / vpn ip / public ip).
2. Ask Admin About port (443).
3. Place certificate files on Chat `Client.app\Content\Resources`.
4. Get `secret.jpeg` image from server admin .
5. Get Coordinates from server admin it will be like this `258,1129;1183,776;972,364;836,194;1320,1085;1431,278;551,715;1335,962;1060,766;1016,268`
6. Get your hash by run `secret.go` . Provide the coordinates generated in the previous step , image name and path make sure the imag will be the same size the same name .
    ```sh
    go run secret.go -image="images/secret.jpeg" -coords="10,10;20,20;30,30"
    ```
    - output will be `Generated Hash: 6de914e9a82db75b457102b217832b5763d9d747e79e9b0b1797012c517dab80
      `.
7. Now Open the app Put your data 
8. Ask Server admin to send your session key 
9. Ask Admin to send you 50-digit number to decrypt the text

**Note:** App only support mac And arm only

