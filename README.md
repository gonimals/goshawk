# Goshawk

Initial prompt:

Create a golang project to monitor servers. The program should be called through command line with the following features:
- Expose a web endpoint to be monitored and which returns "check ok" on every GET request
  - This endpoint should be able to consume a passphrase sent through a GET parameter to confirm who is online on the other side
- The program should parse a configuration JSON-based to define which local or remote services should check. Every service should have one action assigned between ping, web request (with URL, method and expected status) or bash script (with code and expected output through regexp)
  - The configuration file should be validated against a SHA256 provided through command line parameter
  - In case the configuration cannot be validated with the hash, the program should print the calculated hash and work without commands support
- Every time the command detects a service is not working, it should invoke an HTTP request to an endpoint defined in the JSON configuration
- The project module should be github.com/gonimals/goshawk

## Development

### Github workflow

To run the github workflow locally, you can just run this command:
```
act
act push -W .github/workflows/release.yml --container-architecture linux/amd64 \
  --env GORELEASER_CURRENT_TAG=v0.0.1-local --artifact-server-path  ${PWD}/dist #/tmp/artifacts
```

To deploy `act` in an Arch Linux system, just run:

```
run0 pacman -S act goreleaser podman
systemctl --user enable --now podman.socket
act # The medium size image is enough
```