ARCH ?= $(shell arch)

guest: .build/linux/guest/arch/$(ARCH)/boot/bzImage .build/initrd.cpio.gz

.PHONY: guest

.build/linux/guest/arch/$(ARCH)/boot/bzImage: .build/linux/guest/.config
	$(MAKE) -C lib/linux O=$(CURDIR)/.build/linux/guest bzImage

.build/linux/guest/.config: etc/linux/guest/$(ARCH).config
	mkdir -p $(dir $@)
	rsync -c $< $@

menuconfig-guest: .build/linux/guest/.config
	$(MAKE) -C lib/linux O=$(CURDIR)/.build/linux/guest menuconfig
	rsync -c $< etc/linux/guest/$(ARCH).config

.PHONY: menuconfig-guest

.build/initrd.cpio.gz: .build/initrd.iid
	docker save $(shell cat $<) | go run ./cmd/docker2cpio | gzip > $@

.build/initrd.iid: etc/initrd/Dockerfile etc/initrd/init.sh
	mkdir -p .build; docker build --iidfile $@ etc/initrd
