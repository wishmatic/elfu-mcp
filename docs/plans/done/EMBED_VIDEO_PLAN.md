# Inline video playback

Status: Done
Depends on: none

## Goal

Let an agent put a video into the conversation so the user can play it inline in LibreChat, via a new `embed` tool.

`inline` cannot do this. It returns an MCP image block, and both the model and the chat client only ever see image
bytes. A video has to be rendered by the browser, so the chat client has to be handed HTML it is willing to run, which
means an MCP UI resource.

## Non-goals

- Fetching the video, reading it from disk, or proxying it. The user's browser loads the URL itself.
- Deciding whether a URL is a video. There is no extension list, no media type sniffing, and no `type` parameter: the
  caller decides, and any string is embedded as-is.
- Storing or serving anything. This service stays read-only, which is why `embed` does not use the resolver at all.
- Playback that survives the chat's authentication. See "Sandboxed, so unauthenticated" below.

## Design

### The tool

`embed` takes one required argument, `video_url`, and returns two content blocks:

1. a text caption naming the URL, which is what the model reads;
2. an embedded resource whose `uri` starts with `ui://`, whose `mimeType` is `text/html`, and whose `text` is a
   self-contained HTML page holding a single `<video>` element pointed at the URL.

There is no fetch, so `embed` is registered even when no resolver is configured, unlike `inline`.

### Why LibreChat renders it

LibreChat's tool-result parser (`packages/api/src/mcp/parsers.ts`) copies every `resource` item whose URI starts with
`ui://` into a UI-resources artifact, keyed by `sha256(resource.text)` truncated to 10 hex characters, and tells the
model the marker `\ui{<id>}` to place inline. The client keeps those resources addressable for the whole conversation
and rewrites the marker to a renderer, which accepts the resource only when `mimeType` split on `;` equals `text/html`.

Two consequences shape the implementation:

- The HTML is the identity. Because the id hashes the resource text, two different URLs must produce different text;
  they do, since the URL is in both the link and the caption. The same URL embedded twice yields the same id and
  therefore reuses one resource.
- The marker is emitted by the model, not by this service. Nothing here needs to emit it, and the caption should not
  try to.

### The URI

`ui://elfu-mcp/video/<sha256(video_url)[:12]>`, hex.

The URI is only a prefix check for LibreChat, but a distinct resource deserves a distinct identifier, so a client that
keys by URI rather than by the hashed text cannot collide two videos.

### The HTML

One `<video>` with `controls`, `playsinline`, and `preload="metadata"`; the last keeps the chat from pulling whole files
before the user presses play. No `autoplay` and no `muted`, so playback starts on a click, with sound, which is what a
user who asked to see a video expects.

The URL is escaped with `html.EscapeString` before it goes into the `src` attribute. The sandbox does not make an
injection here harmless: whoever asked for the embed controls the URL, and the page runs in the user's browser.

Styling is limited to zeroing the body margin and letting the video fill the iframe width, because LibreChat's renderer
sizes the iframe from the document's content.

### Sandboxed, so unauthenticated

LibreChat renders the resource in a sandboxed iframe with no same-origin permission, so the embedded document has an
opaque origin: it sends no cookies and carries no LibreChat credentials. A URL that only resolves for a signed-in
LibreChat user, or that is only reachable from this service's network, will not play. The URL has to be one the browser
can load on its own.

This is the deliberate trade for not fetching: binding the media to the service would mean downloading and re-serving
videos, which is a file-hosting feature this plan does not build.

## Implementation units

### Unit 1: the presentation

Deliverables: `internal/present/video.go`, `internal/present/video_test.go`.

Acceptance criteria:

- [x] `EmbedVideo(url)` returns exactly two blocks: a text caption naming the URL, then an embedded resource.
- [x] The resource's URI starts with `ui://` and differs per URL, and is byte-identical for a repeated URL.
- [x] The resource's `mimeType` is `text/html`.
- [x] The resource's text is a complete HTML document with one `<video>` element whose `src` is the URL, with
      `controls` and `playsinline`, and without `autoplay` or `muted`.
- [x] A URL containing `&`, `"`, `<`, or `>` is escaped, and the raw sequence never appears in the HTML.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.

### Unit 2: the tool

Deliverables: `internal/mcp/embed.go`, `internal/mcp/embed_test.go`, `internal/mcp/server.go`, and the tool-list
expectations in `internal/mcp/server_test.go` and `internal/server/http_test.go`.

Acceptance criteria:

- [x] `embed` is registered whether or not a resolver is configured, and `tools/list` is
      `[embed, inline, since, sleep, time]` with one and `[embed, since, sleep, time]` without.
- [x] A call with a video URL returns a caption and an embedded resource with a `ui://` URI and `text/html`, and
      `inline`'s behaviour is untouched.
- [x] A call with an empty `video_url` is a tool error, not a panic and not an embed of an empty URL.
- [x] The call makes no network request.
- [x] Over HTTP, with the API key, `embed` returns the same two blocks.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.

### Unit 3: documentation

Deliverables: `README.md`, `docs/PROMPT.md`.

Acceptance criteria:

- [x] `README.md` documents `embed`, that it fetches nothing, and that the URL must be one the browser can load
      without the chat's credentials.
- [x] `docs/PROMPT.md` tells the agent to use `embed` when the user wants to see a video, and to ask for a direct URL
      rather than passing a page URL.
- [ ] Human check: in a real LibreChat, embed a public video URL and confirm the player appears inline and plays.
- [ ] Human check: confirm the model emits the `\ui{...}` marker without being told to, and that a second turn still
      renders the player.

## Verification

Per unit, `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1`. Everything except LibreChat's own
rendering is testable in-process: the content blocks, the escaping, the URI stability, and the tool list. Whether
LibreChat draws the player is the Unit 3 human check, since it needs a browser, an agent, and a model that echoes the
marker.
