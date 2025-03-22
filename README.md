# drateR-go

**drateR** is a simple SFTP server in Go.

## Prerequisites

- Go 1.16 or higher

---
## Download
  Clone the repository to your local machine:

```sh
git clone https://github.com/zeusssz/drateR.git
cd drateR/src
```


## Setup

###

### 1. Install Go Packages

Make sure you have the required Go packages installed:

```sh
go get github.com/pkg/sftp
go get golang.org/x/crypto/ssh
```

### 2. Generate SSH Key Pair

Generate an SSH key pair if you don't already have one:

```sh
ssh-keygen -t rsa -b 2048 -f id_rsa -N ""
```

- **`id_rsa`**: Your private key.
- **`id_rsa.pub`**: Your public key (not needed in this script but required for SSH setup).

### 3. Set Up Environment Variables

Set the `SFTP_PASSWORD` environment variable with your chosen password:

```sh
export SFTP_PASSWORD="your_password"
```

### 4. Configure Private Key Path

Ensure the `privateKeyPath` constant in `drateR.go` is set to the path of your private key:

```go
privateKeyPath = "id_rsa" // Path to the private key for authentication
```

### 5. Run the Server

Navigate to the `src` directory and run the server:

```sh
cd src
go run drateR.go
```

The server will start listening on port `2022` by default.

## Usage

### Connecting to the Server

You can connect to the SFTP server using any SFTP client. Here are a few examples:

#### Using Command Line

```sh
sftp -P 2022 user@<IP_ADDRESS_OF_SERVER>
```

Replace `<IP_ADDRESS_OF_SERVER>` with the IP address of the machine running the SFTP server. Use the password you set in the `SFTP_PASSWORD` environment variable when prompted.

#### Using PuTTY’s `psftp` (Windows)

1. Download PuTTY from [PuTTY Download Page](https://www.chiark.greenend.org.uk/~sgtatham/putty/latest.html).
2. Open `psftp` and connect:

   ```sh
   psftp -P 2022 user@<IP_ADDRESS_OF_SERVER>
   ```

3. Enter the password when prompted.

## Configuration

- **Port**: The server listens on port `2022` by default. Modify the `port` constant in `drateR.go` to change this.
- **Root Directory**: The directory where files are stored is specified by `rootDir`. Change it if needed.

## Troubleshooting

- **Failed to Connect**: Ensure that the server is running and the port is open. Verify the private key path and permissions.
- **Authentication Issues**: Double-check the environment variable `SFTP_PASSWORD` for correct password value and spelling.
