# eSTK.me Cloud Enhance Server

### Introduction

This is a simple server designed to handle requests from the eSTK.me removable eUICC, such as downloads and notifications.

The project is written in Go and is just a toy project, not suitable for production use.

If you want to deploy your own eSTK.me Cloud Enhance server, I recommend you to use the [official eSTK.me Cloud Enhance server](https://github.com/estkme-group/lpac/blob/main/src/rlpa-server.php) instead.

### Installation

You can download the binary release from the [releases page](https://github.com/damonto/estkme-cloud/releases) or you can build it yourself.

If you want to build it yourself, you can run the following commands:
```bash
git clone git@github.com:damonto/estkme-cloud.git
cd estkme-cloud
go build -trimpath -ldflags="-w -s" -o estkme-cloud main.go
```

You can also install the latest version using the following command:

```bash
go install github.com/damonto/estkme-cloud@latest
```

Sometimes, you might need to set executable permissions for the binary file using the following command:

```bash
chmod +x estkme-cloud
```

You must also install the following dependencies:

```bash
# Debian
apt-get install -y --no-install-recommends ca-certificates libpcsclite1 libcurl4
# Arch Linux
pacman -S pcsclite
```

Once done, you can run the server using the following command:

```bash
./estkme-cloud
```

If you want to change the default port, lpac version or data directory, you can use the following flags:

```plaintext
./estkme-cloud --help

Usage of estkme-cloud:
  -data-dir string
        data directory (default "/home/user/workspace/estkme-cloud/data")
  -listen-address string
        eSTK.me cloud enhance server listen address (default ":1888")
  -lpac-version string
        lpac version (default "v2.0.0-beta.1")
  -dont-download
        don't download lpac
  -verbose
        verbose mode
```

If you wish to run the program in the background, you can utilize the systemctl command. Here is an example of how to achieve this:

1. Start by creating a service file in the /etc/systemd/system directory. For instance, you can name the file estkme-cloud.service and include the following content:

```plaintext
[Unit]
Description=eSTK.me Cloud Enhance Server
After=network.target

[Service]
Type=simple
User=your_user_here
Restart=on-failure
ExecStart=/your/binary/path/here/estkme-cloud
RestartSec=10s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```
2. Then, use the following command to start the service:

```bash
systemctl start estkme-cloud
```

3. If you want the service to start automatically upon system boot, use the following command:

```bash
systemctl enable estkme-cloud
```

#### Docker

You can also run the server using Docker. You can use the following command to run the server:

```bash
docker run -d --name estkme-cloud -p 1888:1888 damonto/estkme-cloud:latest
# or use the GitHub Container Registry
docker run -d --name estkme-cloud -p 1888:1888 ghcr.io/damonto/estkme-cloud:latest
```

### Usage

Once the server is running, the server will listen on the specified port (default: 1888) and you can send requests to the server.

#### Download

To download a profile, you should enable the `Cloud Enhance` feature on the eUICC and set the server listening address to the server address. The server will handle the download request and download the profile to the eUICC.

The server support downloading profiles with confirmation code and custom IMEI.

If your eSIM provider requires a confirmation code, you can use the following activation code format:

```plaintext
LPA:1$SM-DP+$Matching Id$$<confirmation_code>
```

Please replace `<confirmation_code>` with the actual confirmation code.

If your eSIM provider requires a custom IMEI, you can use the following activation code format:

1. With confirmation code:
```plaintext
LPA:1$SM-DP+$Matching Id$$<confirmation_code>#<custom_imei>
```

2. Without confirmation code:
```plaintext
LPA:1$SM-DP+$Matching Id#<custom_imei>
```

Please replace `<custom_imei>` with the actual IMEI.

#### Notification

If you click the "Process Notification" button on the eSTK.me eUICC, the server will receive a notification request and send all notifications.

Please note that:

1. All enable, disable and install notifications will be deleted after sending.

2. The delete notifications will be kept in your eSTK.me eUICC.
