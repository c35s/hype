ARCH ?= $(shell arch)

.build/linux/guest/arch/$(ARCH)/boot/bzImage: .build/linux/guest/.config
	$(MAKE) -C lib/linux O=$(CURDIR)/.build/linux/guest bzImage

.build/linux/guest/.config: etc/linux/guest/$(ARCH).config
	mkdir -p $(dir $@)
	rsync -c $< $@

menuconfig-guest: .build/linux/guest/.config
	$(MAKE) -C lib/linux O=$(CURDIR)/.build/linux/guest menuconfig
	rsync -c $< etc/linux/guest/$(ARCH).config

.PHONY: menuconfig-guest
