# gargoyle

HTTP Web Server

## Features

- Reverse Proxying
    - Load Balancing
    - Active Healthchecks
- File Server
    - Browsing Directories
    - Serving Static Sites
- HTTP Response Header Modification
- Authentication
    - Basic HTTP Auth
    - Key Auth
- Optional TLS Support
- URL Rewrites

## Getting Started

**Pre-requisites**
- [Go >= 1.18](https://go.dev/)
- [GNU Make](https://www.gnu.org/software/make/)

### Configuration
- See [CONFIG.md](./CONFIG.md)

### Build from Source

After cloning the repo, run

```shell
make build
```

This will create a binary in the `bin` directory.

### Usage

```shell
./gargoyle <optional path to config file>
```
Default config file path: `./config.json`

## Author
- Kunal Kundu - [@tinfoil-knight](https://github.com/tinfoil-knight)

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.

## Acknowledgements

- Blog Posts & Documentation from
  - [Caddy](https://caddyserver.com/)
  - [Nginx](https://www.nginx.com/)
- Learning Centers
  - [Cloudflare](https://www.cloudflare.com/en-in/learning/)
  - [Kong](https://konghq.com/learning-center)
- Talk on "Building a proxy server in Golang" by [@mauricio](https://github.com/mauricio)