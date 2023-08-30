# Builds and pushes all the images
BUILD_TARGETS = build-ipa-images build-ironic build-qemu-img

all: $(BUILD_TARGETS)

build-%: FORCE
	cd $* && ./push

FORCE:
