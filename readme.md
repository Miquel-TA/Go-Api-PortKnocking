# Port Knocking Web Server

This project is a Go-based web server that uses port knocking to whitelist or deny client IPs from accessing port 80. It listens on a range of ports and requires a specific sequence of port hits (knocks) to allow access to the HTTP server.

## Features

- **Port Knocking:** A sequence of TCP port hits must be made in the correct order to whitelist an IP address.
- **IP Whitelisting:** Successfully knocking in the correct sequence grants access to port 80 for a set period (60 minutes).
- **Dynamic Port Monitoring:** Monitors a range of ports for knock sequences.
- **HTTP Server:** Provides a simple HTTP response to whitelisted IPs.

## How It Works

1. **Port Monitoring:** The server listens on ports 45000 to 45099. When a connection is made, the port number is checked against the expected knock sequence.
2. **Knock Processing:** If the knock sequence is correct, the client's IP is whitelisted for 60 minutes.
3. **Access Control:** Access to the HTTP server on port 80 is granted only to IPs that have successfully completed the knock sequence within the whitelist period.

## Configuration

- **Knock Sequence:** The sequence of ports required to whitelist an IP is defined in the `portKnockSequence` variable. Modify this sequence as needed.
- **Whitelist Duration:** The duration for which an IP is whitelisted is set to 60 minutes. Adjust the `whitelistExpiration` value in the `processKnock` function if a different duration is needed.

