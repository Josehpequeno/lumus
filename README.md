# Lumus

Lumus is a command-line tool written in Go that allows you to read PDF files directly in your terminal. It provides a minimalistic and distraction-free reading experience, making it ideal for reading documents in text format.

## Features

- Read PDF files directly in the terminal
- Navigate through pages easily
- Supports multiple languages for OCR (Optical Character Recognition)
- Minimalistic and distraction-free interface

## Installation

### Prerequisites

Before installing Lumus, you need to have the following dependencies installed on your system:

- Go (version 1.16 or later)
- Tesseract OCR engine
- PDFcpu
- Python3
- Python3-pip
- Python-pypdf2	

### Installation Steps

1. Clone this repository:

   ```bash
   git clone https://github.com/yourusername/lumus.git
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

To use Lumus, simply run the executable followed by the path to the PDF file you want to read:

```bash
lumus path/to/your/file.pdf
```

Once Lumus is running, you can navigate through pages using the arrow keys and perform various actions using the keyboard shortcuts displayed on the screen.

For a complete list of keyboard shortcuts and commands, press `?` while Lumus is running.

## Contributing

Contributions are welcome! If you find any bugs or have suggestions for new features, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Lumus was inspired by [project-name], [link-to-project].

Special thanks to [contributor-names] for their contributions to the project.

---

Você pode personalizar este README conforme necessário, adicionando mais detalhes sobre o projeto, instruções de uso avançado, exemplos de código, etc.
