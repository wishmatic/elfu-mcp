# LFU MCP

> Librechat File Utilities, LFU, or Elfu.

An MCP server with exactly one tool: `inline`. It fetches an image from a URL and returns it as an MCP image block,
downscaled and re-encoded to WebP, so a vision-capable model can look at an image it was only given a link to.

It also serves the stored files it knows about, at `<PUBLIC_HOST>/i/...`.

Built from [Neo MCP](https://github.com/wishmatic/neo-mcp), from which the whole `inline` tool is taken.

## Usage

Deploy as a Docker image:

```sh
docker run -d \
  -p 8080:8080 \
  -e API_KEY=change-me \
  -e PUBLIC_HOST=http://192.168.1.10:8080 \
  -v /mnt/user/appdata/elfu-mcp:/data \
  ghcr.io/wishmatic/elfu-mcp:latest
```

The MCP endpoint is served at `/mcp`; stored images are served from `/i/`.

`/data` holds the file store, so bind-mount a host directory there to keep files across container replacements; the
container runs as uid 65532, so that directory must be writable by it.

See [.env.example](.env.example) for all configuration.

### Authentication

`API_KEY` is required on every `/mcp` request, sent as `Authorization: Bearer <API_KEY>`. Stored images are served
without authentication.

### Tool

- `inline` takes an `image_url` and returns the image inline. Pass any URL the user gave you: the service follows
  redirects, and a URL on this service's own `PUBLIC_HOST` is read straight from the file store instead of over the
  network.

## License

Elfu MCP is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
