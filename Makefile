.PHONY: build-and-setup
build-and-setup:
	chmod +x ./scripts/build-and-setup.sh && ./scripts/build-and-setup.sh

.PHONY: uninstall
uninstall:
	chmod +x ./scripts/uninstall.sh && ./scripts/uninstall.sh

.PHONY: release
release:
	chmod +x ./scripts/release.sh && ./scripts/release.sh
