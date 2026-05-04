# Goshawk

<img src="./docs/logo.svg" alt="Goshawk logo" width="200" height="200">

Goshawk is a fast and lightweight server monitoring tool written in Go. It supports both active and passive monitoring, offering flexible alerting capabilities. Main features include:

- **Passive Monitoring (Heartbeats):** Exposes an HTTP endpoint that returns "check ok" on every GET request.
  - Clients can send a passphrase via a `key` query parameter (`?key=your_secret`) to authenticate and register as active.
  - If an authenticated host stops checking in within a configured timeframe, Goshawk flags it as offline and triggers a notification.
- **Active Monitoring:** Parses a YAML configuration file to define scheduled checks against local or remote services. Supported check types:
  - `tcp`: Attempts a TCP connection to an address/port with a configurable timeout.
  - `web_request`: Sends HTTP requests with customizable method, body, and timeout, validating against an expected status code.
  - `bash_script`: Executes an arbitrary bash script and validates the output using a regular expression.
- **Robust Alerting:** Triggers customizable HTTP POST notifications (e.g., Webhooks, Slack, Telegram) when a service or host state changes.
  - Features configurable retry logic (`max_fails` before alerting) and notification rate-limiting.
  - Notification payloads can be customized using Go templates for titles and bodies.
- **Security:** The configuration file can be securely validated against a SHA256 checksum provided via a command-line argument.

## Installation

The intended way to deploy Goshawk is to run the installer:

```bash
curl -L https://github.com/gonimals/goshawk/raw/refs/heads/main/deploy/install.sh | sudo bash
```

The script downloads the executable file for your architecture and places it in `/usr/local/sbin/goshawk`. It also creates the configuration file at `/etc/goshawk.yml` and installs the systemd service at `/etc/systemd/system/goshawk.service`.

You can also deploy by yourself by by compiling the executable file or downloading from the github releases section. Then, check the installation script to figure out how to run it or just do it manually.

## Configuration

Check the example configuration file to discover how to configure Goshawk. The default configuration file is placed at `/etc/goshawk.yml`, owned by `root` and with permissions `0600` to prevent users from modifying it or reading API keys.

If notifications are not working as expected, it can be a problem with the templates. Errors in templates can be silent, so verify them carefully.

## Development

### Golang

Ensure you have installed `go` in your system.

To keep the project maintained, when developing use these commands:

```bash
go mod tidy
go get -u ./...
go test -cover ./...
go test -race ./...
```

### Github workflow

To run the github workflows locally:

```bash
act push -W .github/workflows/ci.yml --container-architecture linux/amd64 \
  --env GORELEASER_CURRENT_TAG=v0.0.1-local
```

To check what the `goreleaser` tool is generating locally:

```bash
goreleaser release --snapshot --skip=publish --clean
```

To deploy `act` and `goreleaser` in an Arch Linux system:

```bash
run0 pacman -S act goreleaser podman
systemctl --user enable --now podman.socket
act # The medium size image is enough
```
