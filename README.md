# github-widget

[![CI](https://github.com/kirasync2748/github-widget/actions/workflows/ci.yml/badge.svg)](https://github.com/kirasync2748/github-widget/actions/workflows/ci.yml)
[![Docker](https://github.com/kirasync2748/github-widget/actions/workflows/docker.yml/badge.svg)](https://github.com/kirasync2748/github-widget/actions/workflows/docker.yml)
![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Beautiful, embeddable SVG cards for any GitHub repository.**

A small, fast Go API with no frontend, no database and no external Go dependencies. Give it `owner/repo` and it returns a self-contained SVG you can drop into a README, docs site or blog post.

```markdown
![crazy-downloader](https://your-domain.com/api/widget/kirasync2748/crazy-downloader)
```

---

## Features

- **Owner avatar**, repo title, primary language, license and description
- **Colored stats**: stars (gold), forks (blue), open issues (green)
- **Last updated** in plain words ("13 hours ago")
- **Language bar** with official GitHub Linguist colors for all detected languages
- **One-row legend** that always fits, with smart name shortening
- **10 built-in themes**
- **In-memory cache** with TTL and size limits
- **Safe output**: escaped text, no scripts, no `foreignObject`
- Tiny `scratch`-based Docker image (~10 MiB), runs as non-root

---

## Quick Start

```bash
git clone https://github.com/kirasync2748/github-widget.git
cd github-widget
cp .env.example .env     # optionally add GITHUB_TOKEN
docker compose up -d --build
```

---

## Usage

```
GET /api/widget/{owner}/{repo}
```

### Markdown

```markdown
![My Repo](https://your-domain.com/api/widget/kirasync2748/github-widget)
```

### Markdown with a link to the repo

```markdown
[![My Repo](https://your-domain.com/api/widget/kirasync2748/github-widget?theme=dracula)](https://github.com/kirasync2748/github-widget)
```

### HTML

```html
<img src="https://your-domain.com/api/widget/kirasync2748/github-widget?theme=nord&width=600" alt="github-widget" />
```

The response is `Content-Type: image/svg+xml; charset=utf-8` and is cached for 10 minutes (`Cache-Control: public, max-age=600`).

### Query Parameters

| Parameter | Default | Range | Description |
|-----------|---------|-------|-------------|
| `theme`   | `dark`  | see [Themes](#themes) | Color palette |
| `width`   | `540`   | 400–1200 | Card width in pixels |
| `height`  | `200`   | 160–600  | Minimum card height; the card grows to fit its content |

An unknown theme returns `400 Bad Request`.

---

## Themes

Pick one with `?theme=<name>`.

| Theme | Background | Accent | Style |
|-------|------------|--------|-------|
| `dark` _(default)_ | `#111214` | `#58a6ff` | Clean, high-contrast dark |
| `light`    | `#ffffff` | `#0969da` | Bright and minimal |
| `github`   | `#ffffff` | `#0969da` | GitHub's own light look |
| `midnight` | `#0a0e27` | `#7c3aed` | Deep navy with violet |
| `neon`     | `#0d0221` | `#ff00ff` | Vivid purple and pink |
| `ocean`    | `#011627` | `#82aaff` | Deep sea blue |
| `sunset`   | `#1a0f0a` | `#ff7847` | Warm brown and orange |
| `forest`   | `#0f1a0f` | `#56c568` | Dark green, natural |
| `dracula`  | `#282a36` | `#bd93f9` | The classic Dracula palette |
| `nord`     | `#2e3440` | `#81a1c1` | Cool arctic grays |

Stat icon colors stay the same in every theme so they're easy to recognize: stars `#E3B341`, forks `#58A6FF`, issues `#3FB950`.

---

## Language Bar

Language data comes from GitHub's Linguist API, and colors come from [ozh/github-colors](https://github.com/ozh/github-colors).

- **All detected languages** are shown, largest first
- Each one gets a segment in the bar and an entry in the legend
- The legend is **always a single row**:
  1. Full names are used when they fit (`TypeScript 60.4%`)
  2. Otherwise, names longer than 7 characters are shortened (`TypeScr… 60.4%`)
  3. On very narrow cards the text gets a little smaller and names shorter
- Repositories without any code show **"No language data"**
- Unknown languages fall back to `#cccccc`

---

## Configuration

All settings come from environment variables. Copy the template and edit it:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server listen port |
| `GITHUB_TOKEN` | _(empty)_ | Optional token for higher API rate limits |
| `GITHUB_API_URL` | `https://api.github.com` | Change only for GitHub Enterprise |
| `CACHE_TTL` | `10m` | How long cached cards live |
| `CACHE_MAX_ENTRIES` | `1000` | Maximum cached entries |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |

### GitHub Token

Without a token you get **60 requests/hour** per IP. With one you get **5,000 requests/hour**. No scopes are needed for public repositories.

<details>
<summary><b>How to create a token</b></summary>

**Classic token**

1. Go to <https://github.com/settings/tokens>
2. Click **Generate new token → Generate new token (classic)**
3. Name it (e.g. `github-widget`)
4. Leave all scopes unchecked
5. Click **Generate token** and copy it (starts with `ghp_`)

**Fine-grained token**

1. Go to <https://github.com/settings/personal-access-tokens/new>
2. Set a name and expiration
3. Under **Repository access** choose **Public Repositories (read-only)**
4. Generate and copy it (starts with `github_pat_`)

Then set it in `.env`:

```bash
GITHUB_TOKEN=ghp_your_token_here
```

</details>

The token never appears in SVG output, logs or error responses.

---

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | JSON with `status`, `ping`, `version` and `time` |
| `GET /healthz` | Liveness check: `{"status":"ok"}` |
| `GET /readyz` | Readiness check: `{"status":"ready"}` |
| `GET /api/widget/{owner}/{repo}` | The SVG card |

| Error | Status |
|-------|--------|
| Invalid owner/repo or unknown theme | `400` |
| Repository not found | `404` |
| GitHub rate limit reached | `429` |
| GitHub upstream error | `502` |

---

## Development

Requires Go 1.27 (or just Docker).

```bash
cp .env.example .env
make run                 # or: go run ./cmd/server
```

The server runs at <http://localhost:3000>.

### Docker Compose (live reload)

`docker-compose.yml` mounts the source into a Go container and runs `go run`, so a restart picks up your changes, with no rebuild needed:

```bash
docker compose up -d --build      # or: make compose-up
docker compose restart server     # apply code changes
docker compose down               # or: make compose-down
```

### Tests

```bash
make test        # go test ./...
make test-race   # with the race detector
make verify      # gofmt check + vet + tests + race + build
```

### Versioning

The version defaults to `dev` and is set at build time:

```bash
make build VERSION=v1.2.0
# or
go build -ldflags="-X github.com/kirasync2748/github-widget/internal/version.Version=v1.2.0" ./cmd/server
```

It's reported by `GET /`.

---

## Docker

```bash
make docker-build VERSION=v1.2.0
docker run --rm -p 3000:3000 -e GITHUB_TOKEN=ghp_xxx github-widget:latest
```

Release images are published to GHCR:

```bash
docker pull ghcr.io/kirasync2748/github-widget:latest
```

---

## Deployment

### Vercel (native Go)

1. Push the repo to GitHub
2. Import it on [Vercel](https://vercel.com)
3. Deploy. Vercel detects the Go project and routes to `$PORT`

### Vercel (Docker)

Use `Dockerfile.vercel`: in **Project Settings → General**, set the build to use the Dockerfile. The container listens on `$PORT`.

### Any container host

Run the GHCR image anywhere that runs containers, and expose port `3000`.

---

## Security

- Owner and repo names are validated against GitHub's naming rules
- Every dynamic string is XML-escaped before it goes into the SVG
- SVG output has no scripts, event handlers or `foreignObject`
- Only the GitHub API is contacted; no arbitrary URL fetching
- The GitHub token is never logged or exposed

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-change`
3. Run `make verify`
4. Open a pull request

