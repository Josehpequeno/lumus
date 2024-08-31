debian:
	 make build && cp lumus bin/lumus && sudo dpkg-buildpackage -b -uc
