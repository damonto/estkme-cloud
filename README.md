# eSTK.me Remote LPA Server

### Introduction

This is a simple server designed to handle requests from the eSTK.me removable eUICC, such as downloads and notifications. The project is written in Go and is just a toy project, not suitable for production use.

If you want to deploy your own rLPA server, I recommend you to use the [official rLPA server](https://github.com/estkme-group/lpac/blob/main/src/rlpa-server.php) instead.


### Installation and Usage

You can download the binary release from the [releases page](https://github.com/damonto/estkme-rlpa-server/releases) or you can build it yourself.

If you want to build it yourself, you can run the following commands:
```bash
git@github.com:damonto/estkme-rlpa-server.git
cd estkme-rlpa-server
go build -trimpath -ldflags="-w -s" -o estkme-rlpa-server main.go
```

If you have already installed Go (Golang) on your system, you can also install the latest version using the following command:

```bash
go install github.com/damonto/estkme-rlpa-server@latest
```

Sometimes, you might need to set executable permissions for the binary file using the following command:

```bash
chmod +x estkme-rlpa-server
```

Once done, you can run the server using the following command:

```bash
./estkme-rlpa-server
```

If you want to change the default port, lpac version or data directory, you can use the following flags:

```plaintext
./estkme-rlpa-server --help

Usage of estkme-rlpa-server:
  -data-dir string
        data directory (default "/home/user/workspace/estkme-rlpa-server/data")
  -listen-address string
        rLPA server listen address (default ":1888")
  -lpac-version string
        lpac version (default "v2.0.0-beta.1")
```

If you wish to run the program in the background, you can utilize the systemctl command. Here is an example of how to achieve this:

1. Start by creating a service file in the /etc/systemd/system directory. For instance, you can name the file telegram-sms.service and include the following content:

```plaintext
[Unit]
Description=eSTK.me rLPA Server
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
ExecStart=/your/binary/path/here/estkme-rlpa-server
RestartSec=10s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```
2. Then, use the following command to start the service:

```bash
systemctl start estkme-rlpa-server
```

3. If you want the service to start automatically upon system boot, use the following command:

```bash
systemctl enable estkme-rlpa-server
```
