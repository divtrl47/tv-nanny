# tv-nanny

This repository contains a simple WebOS application that shows a pie clock. The circle is divided into sections defined in `sections.yaml`. A hand shows the current time and highlights the active section.

## Running

1. Copy the `webos-app` directory to your LG TV or run it in a browser that supports WebOS web apps.
2. Open `index.html` to see the clock. The app reads `sections.yaml` at startup.

You can edit `sections.yaml` to configure names, colors and start times.

## Makefile

You can use the Makefile for common tasks:

- `make serve` – start a local web server on port 8000.
- `make package` – create `tv-nanny.zip` with the app content.
- `make clean` – remove the generated archive.
- `make test` – run a simple Node.js check to ensure the app loads without
  ReferenceErrors.
