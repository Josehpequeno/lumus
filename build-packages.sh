#!/bin/bash

set -e
#Esta linha faz com que o script termine imediatamente se qualquer comando falhar (retornar um código de saída não zero).

VERSION="1.0.1"
NAME="lumus"
BUILD_DIR="build"
DIST_DIR="dist"

mkdir -p $BUILD_DIR $DIST_DIR

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/$NAME lumus.go

# Criar tarball
echo "Creating distribution tarball..."
mkdir -p $BUILD_DIR/pkg/usr/bin
mkdir -p $BUILD_DIR/pkg/usr/share/$NAME
#Cria a estrutura de diretórios dentro de build/pkg que simula a estrutura de diretórios do Linux, com usr/bin para o binário e usr/share/lumus para os dados.
cp $BUILD_DIR/$NAME $BUILD_DIR/pkg/usr/bin/
tar -czf $DIST_DIR/$NAME-$VERSION-linux-amd64.tar.gz -C $BUILD_DIR/pkg .
#Cria um arquivo tar compactado (gzip) a partir do conteúdo do diretório pkg (que contém a estrutura usr). A opção -C muda o diretório para $BUILD_DIR/pkg antes de criar o tar, então o tar incluirá os arquivos a partir dali (ou seja, a estrutura inside do tar começará com usr/).

echo "Package created: $DIST_DIR/$NAME-$VERSION-linux-amd64.tar.gz"

# Criar pacote .deb (Debian/Ubuntu)
echo "Creating Debian package..."
mkdir -p $BUILD_DIR/deb/DEBIAN
mkdir -p $BUILD_DIR/deb/usr/bin
mkdir -p $BUILD_DIR/deb/usr/share/$NAME

cat > $BUILD_DIR/deb/DEBIAN/control << EOF
Package: $NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Josehpequeno <hicarojbs21@gmail.com>
Description: A command line tool to read PDF files directly in the terminal.
EOF

cp $BUILD_DIR/$NAME $BUILD_DIR/deb/usr/bin/
dpkg-deb --build $BUILD_DIR/deb $DIST_DIR/${NAME}_${VERSION}_amd64.deb

echo "All packages created in $DIST_DIR/:"
ls -la $DIST_DIR/