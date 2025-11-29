# Variáveis
BINARY_NAME=lumus
VERSION=1.0.1
BUILD_DIR=build
PREFIX=/usr/local
PKGBUILD_SRC=PKGBUILD
PKGBUILD_TEMP=PKGBUILD.temp
NAME=lumus

# Variáveis RPM
RPM_NAME=$(BINARY_NAME)
RPM_VERSION=$(VERSION)
RPM_RELEASE=1
RPM_DIR=rpmbuild
RPM_SOURCE_DIR=$(RPM_DIR)/SOURCES
RPM_SPEC_DIR=$(RPM_DIR)/SPECS
RPM_BUILD_DIR=$(RPM_DIR)/BUILD
RPM_RPMS_DIR=$(RPM_DIR)/RPMS
RPM_SRPMS_DIR=$(RPM_DIR)/SRPMS

DEB_DIR=deb-package

# Targets especiais
#.PHONY: Indica que esses são "alvos falsos" (não são arquivos reais)
.PHONY: all build clean install uninstall pkgbuild clean-pkg arch-package install-arch \
	deb-package prepare-aur update-aur clean-aur clean-all rpm-dirs rpm-spec rpm-tarball rpm-package install-rpm clean-rpm check-rpm-deps rpm

# Target padrão - executa quando você só digita 'make'
all: build

# Compila o programa
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) lumus.go
	@echo "Build complete!"

# Compila para Linux
build-linux:
	@echo "Building $(BINARY_NAME) v$(VERSION) for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux lumus.go
	@echo "Linux build complete!"

# Limpa os arquivos de build
clean:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Instala o programa no sistema
#    install: Comando do Linux para instalar arquivos

#    -D: Cria todos os diretórios necessários

#    -m755: Define permissões (755 = rwxr-xr-x - executável)

install:
	@echo "Installing $(BINARY_NAME)..."
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	@echo "Installation complete!"



# Desinstala o programa
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	rm -rf $(DESTDIR)$(PREFIX)/share/$(BINARY_NAME)
	@echo "Uninstallation complete!"

# Cria o tarball e atualiza o PKGBUILD automaticamente
pkgbuild:
	@echo "Creating source tarball for $(BINARY_NAME)..."
	tar -czf $(NAME)-$(VERSION).tar.gz lumus.go spinner/spinner.go LICENSE go.mod go.sum 
	@echo "Tarball created: $(NAME)-$(VERSION).tar.gz"
	@echo "Run: makepkg -s"

# Limpa os arquivos de empacotamento
clean-pkg:
	@echo "Cleaning package files..."
	rm -f $(NAME)-*.tar.gz
	rm -f *.pkg.tar.*
	@echo "Package clean complete!"

# Alvo completo: limpa, cria tarball e constrói pacote
arch-package: clean clean-pkg pkgbuild
	@echo "Building Arch Linux package..."
	makepkg -s
	@echo "Package build complete!"

# Instala o pacote localmente (para teste)
install-arch: arch-package
	@echo "Installing Arch Linux package..."
	sudo pacman -U $(NAME)-*.pkg.tar.*
	@echo "Package installation complete!"

deb-package: build-linux
	@echo "Creating Debian package..."
	@mkdir -p $(DEB_DIR)/DEBIAN
	@mkdir -p $(DEB_DIR)/usr/bin
	@mkdir -p $(DEB_DIR)/usr/share/doc/lumus
	
	# Copiar binário
	cp $(BUILD_DIR)/lumus-linux $(DEB_DIR)/usr/bin/lumus
	chmod +x $(DEB_DIR)/usr/bin/lumus
	
	# Copiar documentação
	cp LICENSE $(DEB_DIR)/usr/share/doc/lumus/
	
	# Criar arquivo de controle usando echo
	@echo "Package: lumus" > $(DEB_DIR)/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DEB_DIR)/DEBIAN/control
	@echo "Architecture: amd64" >> $(DEB_DIR)/DEBIAN/control
	@echo "Maintainer: Josehpequeno <hicarojbs21@gmail.com>" >> $(DEB_DIR)/DEBIAN/control
	@echo "Depends: libc6 (>= 2.28), libgcc-s1 (>= 3.0), liblept5, libstdc++6 (>= 6), libtesseract5 (>= 4.0.0)" >> $(DEB_DIR)/DEBIAN/control
	@echo "Recommends: poppler-utils, wv, unrtf, tidy, tesseract-ocr" >> $(DEB_DIR)/DEBIAN/control
	@echo "Section: utils" >> $(DEB_DIR)/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_DIR)/DEBIAN/control
	@echo "Homepage: https://github.com/Josehpequeno/lumus" >> $(DEB_DIR)/DEBIAN/control
	@echo "Description: PDF reader for terminal" >> $(DEB_DIR)/DEBIAN/control
	@echo " A command line tool to read PDF files directly in the terminal." >> $(DEB_DIR)/DEBIAN/control
	@echo " Provides a minimalistic and distraction-free reading experience." >> $(DEB_DIR)/DEBIAN/control
	
	# Construir pacote .deb
	dpkg-deb --build $(DEB_DIR) lumus_$(VERSION)_amd64.deb
	@echo "Pacote Debian criado: lumus_$(VERSION)_amd64.deb"

clean:
	rm -rf $(BUILD_DIR) $(DEB_DIR) *.deb

# Comandos para AUR
prepare-aur:
	@echo "🚀 Preparando pacote para AUR..."
	@chmod +x prepare-aur.sh
	@./prepare-aur.sh

# Atualizar AUR (após mudanças)
update-aur: clean
	@echo "🔄 Atualizando AUR..."
	@mkdir -p lumus-$(VERSION)
	@cp lumus.go lumus-$(VERSION)/
	@tar -czf lumus-$(VERSION).tar.gz lumus-$(VERSION)/
	@rm -rf lumus-$(VERSION)
	@if [ -d "aur" ]; then \
		cp lumus-$(VERSION).tar.gz aur/; \
		cd aur && makepkg --printsrcinfo > .SRCINFO; \
		echo "✅ AUR atualizado localmente"; \
		echo "💡 Agora faça:"; \
		echo "   cd aur && git add . && git commit -m 'Update to v$(VERSION)' && git push"; \
	else \
		echo "❌ Diretório aur não encontrado. Execute 'make prepare-aur' primeiro"; \
	fi

# Limpar também arquivos AUR
clean-aur:
	rm -rf aur lumus-*.tar.gz

clean-all: clean clean-aur


# =============================================================================
# TARGETS RPM CORRIGIDOS
# =============================================================================

# Cria a estrutura de diretórios para RPM
rpm-dirs:
	@echo "Creating RPM directory structure..."
	mkdir -p $(RPM_SOURCE_DIR)
	mkdir -p $(RPM_SPEC_DIR)
	mkdir -p $(RPM_BUILD_DIR)
	mkdir -p $(RPM_RPMS_DIR)
	mkdir -p $(RPM_SRPMS_DIR)

# Cria o arquivo .spec para RPM
rpm-spec: rpm-dirs
	@echo "Creating RPM spec file..."
	@echo "Name: $(RPM_NAME)" > $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "Version: $(RPM_VERSION)" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "Release: $(RPM_RELEASE)%{?dist}" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "Summary: Display random cute ASCII art emojis" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "License: MIT" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "URL: https://github.com/Josehpequeno/lumus" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "Source0: %{name}-%{version}.tar.gz" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "BuildRequires: golang" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "BuildRequires: make" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%description" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "A simple Go program that displays random cute ASCII art emojis from a collection." >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%prep" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%setup -q" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%build" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "make build" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%install" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "mkdir -p %{buildroot}/usr/bin" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "mkdir -p %{buildroot}/usr/share/lumus" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "install -m 755 build/lumus %{buildroot}/usr/bin/lumus" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%files" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "/usr/bin/lumus" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "/usr/share/lumus/*" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "%changelog" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "* $(shell LANG=C date '+%a %b %d %Y') Josehpequeno <hicarojbs21@gmail.com> - $(RPM_VERSION)-$(RPM_RELEASE)" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "- Initial package" >> $(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "RPM spec file created: $(RPM_SPEC_DIR)/$(RPM_NAME).spec"

# Cria o tarball fonte para RPM
rpm-tarball: rpm-dirs
	@echo "Creating source tarball for RPM..."
	mkdir -p $(NAME)-$(VERSION)
	cp lumus.go $(NAME)-$(VERSION)/
	cp Makefile $(NAME)-$(VERSION)/
	[ -f LICENSE ] && cp LICENSE $(NAME)-$(VERSION)/ || echo "LICENSE not found, continuing..."
	[ -f go.mod ] && cp go.mod $(NAME)-$(VERSION)/ || echo "go.mod not found, continuing..."
	tar -czf $(RPM_SOURCE_DIR)/$(NAME)-$(VERSION).tar.gz $(NAME)-$(VERSION)/
	rm -rf $(NAME)-$(VERSION)
	@echo "Source tarball created: $(RPM_SOURCE_DIR)/$(NAME)-$(VERSION).tar.gz"

# Constrói o pacote RPM
rpm-package: build rpm-spec rpm-tarball
	@echo "Building RPM package..."
	rpmbuild -ba \
		--define "_topdir $(CURDIR)/$(RPM_DIR)" \
		--define "_version $(VERSION)" \
		--define "_release $(RPM_RELEASE)" \
		$(RPM_SPEC_DIR)/$(RPM_NAME).spec
	@echo "RPM package built!"
	@echo "RPM files location: $(RPM_RPMS_DIR)"

# Instala o RPM localmente (para teste)
install-rpm: rpm-package
	@echo "Installing RPM package..."
	@rpm_file=$$(find $(RPM_RPMS_DIR) -name "*.rpm" | head -1); \
	if [ -n "$$rpm_file" ]; then \
		sudo rpm -ivh --force $$rpm_file; \
		echo "RPM installed successfully!"; \
	else \
		echo "No RPM file found!"; \
		exit 1; \
	fi

# Limpa os arquivos RPM
clean-rpm:
	@echo "Cleaning RPM files..."
	rm -rf $(RPM_DIR)
	@echo "RPM clean complete!"

# Verifica se as dependências RPM estão instaladas
check-rpm-deps:
	@echo "Checking RPM build dependencies..."
	@command -v rpmbuild >/dev/null 2>&1 || { \
		echo "Error: rpmbuild is not installed."; \
		echo "On Fedora/RHEL/CentOS: sudo dnf install rpm-build"; \
		exit 1; \
	}
	@command -v go >/dev/null 2>&1 || { \
		echo "Error: golang is not installed."; \
		echo "On Fedora/RHEL/CentOS: sudo dnf install golang"; \
		exit 1; \
	}
	@echo "All RPM dependencies are satisfied."

# Target completo para RPM (com verificação de dependências)
rpm: check-rpm-deps rpm-package
	@echo "🎉 RPM package build complete!"
	@echo "📦 RPM files: $(RPM_RPMS_DIR)"
