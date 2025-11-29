# Lumus

Lumus is a command-line tool written in Go that allows you to read PDF files directly in your terminal. It provides a minimalistic and distraction-free reading experience, making it ideal for reading documents in pdf format.

## Features

- Read PDF files directly in the terminal
- Navigate through pages easily
- Minimalistic and distraction-free interface

## Preview
![Preview](preview.gif)

## Installation

### Prerequisites

Before installing Lumus, you need to have the following dependencies installed on your system:

- Go (version 1.16 or later)

#### Debian/Ubuntu

To install the default version available in the repositories:
```
sudo apt-get install tesseract-ocr libleptonica-dev libtesseract-dev poppler-utils wv unrtf tidy
```

Verify the Tesseract installation:
```
tesseract --version
```

If Tesseract is not version 5, add the PPA and update:
```
sudo add-apt-repository ppa:alex-p/tesseract-ocr5
sudo apt-get update
sudo apt-get install tesseract-ocr
```

#### Arch linux

```
sudo pacman -S tesseract poppler wv unrtf tidy leptonica
```

### Installation Steps

#### Install from AUR (Arch Linux)

You can install Lumus from the AUR using your favorite AUR helper:

**Using yay:**
```bash
yay -S lumus
```

**Using paru:**
```bash
paru -S lumus
```

**Manual installation from AUR:**
```bash
git clone https://aur.archlinux.org/lumus.git
cd lumus
makepkg -si
```

#### Install from Source

**Install on Arch:**
1. Clone this repository:
   ```bash
   git clone https://github.com/Josehpequeno/lumus
   ```
2. Install base-devel:
   ```bash
   sudo pacman -Syu base-devel
   ```
3. Navigate to the project directory:
   ```bash
   cd lumus
   ```
4. Install with makepkg:
   ```bash
   makepkg -si
   ```

**Install on Debian/Ubuntu:**
1. Download file `lumus_1.0.1-1_amd64.deb` from releases
   ```bash
   sudo dpkg -i lumus_1.0.1-1_amd64.deb && sudo apt-get install -f
   ```

## Usage

To use Lumus, simply run the executable and navigate to the file you want to read:

```bash
lumus
```

Once Lumus is running, you can navigate through pages using the arrow keys and perform various actions using the keyboard shortcuts displayed on the screen.

## Contributing

Contributions are welcome! If you find any bugs or have suggestions for new features, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Future Work

The development of a more thematic loading animation are in future plans, in addition to some future user recommendations.
