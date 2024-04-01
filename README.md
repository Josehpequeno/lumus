# Lumus

Lumus is a command-line tool written in Go that allows you to read PDF files directly in your terminal. It provides a minimalistic and distraction-free reading experience, making it ideal for reading documents in pdf format.

## Features

- Read PDF files directly in the terminal
- Navigate through pages easily
- Minimalistic and distraction-free interface

## Installation

### Prerequisites

Before installing Lumus, you need to have the following dependencies installed on your system:

- Go (version 1.16 or later)

# Debian
- tesseract-ocr
- libleptonica-dev
- libtesseract-dev
- python3-pip
- pip install PyPDF2

# Arch linux
- tesseract
- yay -S python-pypdf2


### Installation Steps

1. Clone this repository:

   ```bash
   git clone https://github.com/Josehpequeno/lumus
   ```

2. Navigate to the project directory:

   ```bash
   cd lumus
   ```

3. Build the project:

   ```bash
   go build
   ```

4. Install the executable:

   ```bash
   go install
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

<!-- ## Acknowledgments -->

<!-- Lumus was inspired by [project-name], [link-to-project]. -->

<!-- Special thanks to [contributor-names] for their contributions to the project. -->
