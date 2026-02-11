# openport

Expose your local services to the internet. Like ngrok, but simple and open source.

## How it works

You run the `op` CLI on your machine. It connects to an openport server and gives you a public URL. Any traffic to that URL gets forwarded to your local service.

```
Internet  →  openport server  →  tunnel  →  your machine  →  localhost:3000
```

## Install

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/nitintf/openport/releases/latest/download/op-darwin-arm64 -o op
chmod +x op
sudo mv op /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/nitintf/openport/releases/latest/download/op-darwin-amd64 -o op
chmod +x op
sudo mv op /usr/local/bin/
```

**Linux (amd64)**
```bash
curl -L https://github.com/nitintf/openport/releases/latest/download/op-linux-amd64 -o op
chmod +x op
sudo mv op /usr/local/bin/
```

**From source**
```bash
go install github.com/nitintf/openport/cmd/op@latest
```

## Usage

Start your local server, then expose it:

```bash
op 3000
```

```
  openport v0.1.0

  Forwarding   http://a1b2c3d4.localhost:8080  →  http://localhost:3000

  Press Ctrl+C to stop
  ────────────────────────────────────────────────────────

  ● 15:04:05 200 GET     /               3ms
  ● 15:04:05 200 GET     /styles.css     1ms
  ● 15:04:07 201 POST    /api/users      45ms
```

**Options**

```bash
op 3000                                        # expose port 3000
op 8080 --server tunnel.example.com:9090       # use a custom server
op 4000 --subdomain myapp                      # request a specific subdomain
op --version                                   # print version
```

## Self-hosting the server

If you want to run your own openport server:

```bash
# from a release
curl -L https://github.com/nitintf/openport/releases/latest/download/openport-server-linux-amd64 -o openport-server
chmod +x openport-server
./openport-server

# from source
go run ./cmd/server
```

The server listens on `:8080` for public HTTP traffic and `:9090` for tunnel connections. Configure with flags:

```bash
openport-server -addr :8080 -tunnel-addr :9090 -domain yourdomain.com
```

Point a wildcard DNS record (`*.yourdomain.com`) at your server, and clients can connect with:

```bash
op 3000 --server yourdomain.com:9090
```

## Contributing

```bash
git clone https://github.com/nitintf/openport.git
cd openport
make build        # build both binaries
make test         # run tests
```

The project is structured as:

```
cmd/op/          CLI client
cmd/server/      tunnel server
internal/        shared packages
```

Fork the repo, create a branch, make your changes, and open a pull request.

## License

MIT
