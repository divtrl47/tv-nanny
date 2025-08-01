.PHONY: serve package clean test

serve:
	python3 -m http.server -d webos-app 8000

package:
	cd webos-app && zip -r ../tv-nanny.zip .

clean:
	rm -f tv-nanny.zip

test:
	node test/test.js
