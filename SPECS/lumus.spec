Name: lumus
Version: 1.0.1
Release: 1%{?dist}
Summary: Display random cute ASCII art emojis

License: MIT
URL: https://github.com/Josehpequeno/lumus
Source0: %{name}-%{version}.tar.gz

BuildRequires: golang
BuildRequires: make

%description
A simple Go program that displays random cute ASCII art emojis from a collection.

%prep
%setup -q

%build
make build

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/lumus
install -m 755 build/lumus %{buildroot}/usr/bin/lumus

%files
/usr/bin/lumus
/usr/share/lumus/*

%changelog
* sáb nov 29 2025 Josehpequeno <hicarojbs21@gmail.com> - 1.0.1-1
- Initial package
