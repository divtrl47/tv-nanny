.PHONY: serve package clean test install run

serve:
        go run .

package:
        cd webos-app && zip -r ../tv-nanny.zip .

clean:
        rm -f tv-nanny.zip

install:
        apt-get update
        apt-get install -y docker.io docker-compose-plugin

run:
        docker compose up --build

test:
        go test ./...
