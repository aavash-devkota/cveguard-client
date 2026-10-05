# CVEGuard Client
[![Go](https://img.shields.io/badge/Go-CLI-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go CLI client for the **CVEGuard** vulnerability intelligence platform.
Talks to [cveguard-server](https://github.com/aavash-devkota/cveguard-server) and lets you
query vulnerability data from the terminal.

## Build & run

```bash
go build -o cveguard .
./cveguard --help
```

## Requires

- Go 1.21+
- A running [cveguard-server](https://github.com/aavash-devkota/cveguard-server)

## License

MIT
