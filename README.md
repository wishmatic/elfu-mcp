<img src="docs/images/logo.webp" alt="Elfu MCP Logo" width="128">

# LFU MCP

> Librechat File Utilities, LFU, or Elfu.

An MCP server for looking at images. It has one tool, `inline`: give it an image URL and it returns the image itself as
an MCP image block, downscaled and re-encoded to WebP, so a vision-capable model can see it.

Built from [Neo MCP](https://github.com/wishmatic/neo-mcp), from which the whole `inline` tool is taken.

## The tool

`inline` takes an `image_url`, follows redirects, and returns the image. Pass any URL the user gives you.

An image the user pastes into a chat has no URL you can read, so ask them for one; see
[docs/PROMPT.md](docs/PROMPT.md) for the agent instructions that say so.

## Reading URLs the model cannot fetch

Some of the URLs a model is given are not reachable from this service:

- a LibreChat image URL needs a session cookie, so `inline` cannot download it;
- an image URL that is another MCP's public host, which resolves differently inside the network.

`INLINE_URL_MAP` maps either kind to somewhere this service can read it, as comma-separated `public=private` pairs where
the private side is an absolute `http(s)` base URL or a directory:

```sh
# LibreChat's uploaded images, with its images directory mounted into this container
INLINE_URL_MAP=https://chat.example.com/images/=/data/librechat-data

# Images served by another MCP, reached under a private name
INLINE_URL_MAP=https://neo.example.com=http://neo-mcp:8080
```

A URL under a public key is rewritten to the private side before it is fetched, so the model can pass either kind to
`inline` unchanged.

A directory entry reads a file this service has mounted. Its public key is the only thing standing between a caller and
that directory, so make the key long enough that it cannot be guessed. A key appearing in a conversation or in this
service's logs is expected; the private side is never logged.

## Usage

Deploy as a Docker image:

```sh
docker run -d \
  -p 8080:8080 \
  -e API_KEY=change-me \
  -e PUBLIC_HOST=http://192.168.1.10:8080 \
  -e INLINE_URL_MAP=https://chat.example.com/images/=/data/librechat-data \
  -v /mnt/user/appdata/elfu-mcp:/data \
  -v /mnt/user/appdata/librechat/client/public/images:/data/librechat-data:ro \
  ghcr.io/wishmatic/elfu-mcp:latest
```

The MCP endpoint is served at `/mcp`; stored images are served from `/i/`.

`/data` holds the file store, so bind-mount a host directory there to keep files across container replacements; the
container runs as uid 65532, so that directory must be writable by it.

See [.env.example](.env.example) for all configuration.

### Authentication

`API_KEY` is required on every `/mcp` request, sent as `Authorization: Bearer <API_KEY>`. Stored images are served
without authentication.

## License

Elfu MCP is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
