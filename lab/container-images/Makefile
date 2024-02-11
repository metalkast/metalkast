# Builds and pushes all the images
BUILD_TARGETS = build-ipa-images build-ironic

all: $(BUILD_TARGETS)

build-%: FORCE
	cd $* && ./push

FORCE:
