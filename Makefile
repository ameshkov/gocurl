NAME=gocurl
BASE_BUILDDIR=build
BUILDNAME=$(GOOS)-$(GOARCH)
BUILDDIR=$(BASE_BUILDDIR)/$(BUILDNAME)
VERSION?=v0.0-dev
VERSIONPKG=github.com/ameshkov/gocurl/internal/appversion

ifeq ($(GOOS),windows)
  ext=.exe
  archiveCmd=zip -9 -r $(NAME)-$(BUILDNAME)-$(VERSION).zip $(BUILDNAME)
else
  ext=
  archiveCmd=tar czpvf $(NAME)-$(BUILDNAME)-$(VERSION).tar.gz $(BUILDNAME)
endif

.PHONY: default check init
default: build

build: clean
	CGO_ENABLED=0 go build -ldflags "-X $(VERSIONPKG).version=$(VERSION)" -o $(NAME)

release: check-env-release
	mkdir -p $(BUILDDIR)
	cp LICENSE $(BUILDDIR)/
	cp README.md $(BUILDDIR)/
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X $(VERSIONPKG).version=$(VERSION)" -o $(BUILDDIR)/$(NAME)$(ext)
	cd $(BASE_BUILDDIR) ; $(archiveCmd)

lint:
	npx markdownlint-cli .
	golangci-lint run

test:
	go test -race -v -bench=. ./...

# check runs linters and tests, it mirrors the CI quality gate.
check: lint test

# init installs the git pre-commit hook, see .githooks/pre-commit.
init:
	@ echo "Installing the gocurl pre-commit hook..."
	@ mkdir -p .git/hooks
	@ cp .githooks/pre-commit .git/hooks/pre-commit
	@ chmod +x .git/hooks/pre-commit
	@ echo "Done. The pre-commit hook is installed at .git/hooks/pre-commit."
	@ echo "It runs on every commit: checks for temporary TODOs, linters, tests."

clean:
	go clean
	rm -rf $(BASE_BUILDDIR)

check-env-release:
	@ if [ "$(GOOS)" = "" ]; then \
		echo "Environment variable GOOS not set"; \
		exit 1; \
	fi
	@ if [ "$(GOARCH)" = "" ]; then \
		echo "Environment variable GOARCH not set"; \
		exit 1; \
	fi
