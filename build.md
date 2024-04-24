debian:
	 sudo apt-get update && sudo apt-get install dh-make-golang
  	dh-make-golang && dpkg-buildpackage -b -uc



