all: bin/app

PLATFORM=local

.PHONY: bin/app
bin/app:
	@docker build . --target bin


