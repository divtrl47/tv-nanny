# tv-nanny

A simple schedule clock.

The Go backend reads `config.yaml`, which defines reusable program entries
(`sleep`, `play`, `eat`, `walk`, `shower`) and the day's schedule. It renders a
high‑resolution PNG of the clock with 10‑minute color gradients between adjacent
sections and computes positions for the emoji icons. The frontend is a small
page that polls the `/image` endpoint every five seconds using
[htmx](https://htmx.org) and overlays the emoji at those coordinates.

## Running locally

```sh
go run .
```

Open `webos-app/index.html` in a browser; it requests the rendered image from
the Go server running on port `8080`.

## Docker

Build and run both the Go backend and the nginx frontend:

```sh
docker compose up --build
```

Nginx serves the static `webos-app` on port `80` and proxies `/image` requests
to the Go service.

## Configuration

Edit `config.yaml` to adjust program colors, emoji, or the schedule. Changes are
picked up on the next request.
