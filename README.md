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

#### Debian

```
sudo apt-get install tesseract-ocr libleptonica-dev libtesseract-dev poppler-utils wv unrtf tidy
```

#### Arch linux

```
sudo pacman -S tesseract poppler wv unrtf tidy leptonica
```

### Installation Steps
<!--  -->
<!-- 1. Clone this repository: -->
<!--  -->
   <!-- ```bash -->
   <!-- git clone https://github.com/Josehpequeno/lumus -->
   <!-- ``` -->
<!--  -->
<!-- 2. Navigate to the project directory: -->
<!--  -->
   <!-- ```bash -->
   <!-- cd lumus -->
   <!-- ``` -->
<!--  -->
<!-- 3. Build the project: -->
<!--  -->
   <!-- ```bash -->
   <!-- go build -->
   <!-- ``` -->
<!--  -->
<!-- 4. Install the executable: -->
<!--  -->
   <!-- ```bash -->
   <!-- go install -->
   <!-- ``` -->

- Install on Arch
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
   4. Install with makepkg

      ```
      makepkg -si
      ```
- Install on Debian
   1. Download file lumus_1.0.0-1_amd64.deb on release
      ```
      sudo dpkg -i lumus_1.0.0-1_amd64.deb && sudo apt-get install -f
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

<!-- ## Explanation -->
<!--  -->
<!-- The project uses python code with the PyPDF2 library to extract texts from pages. The gosseract library is used as a complement to PyPDF2, as it extracts text from images in PDFs. The Levenshtein Distance algorithm is used to determine whether the text extracted from the images and the PDF are similar. If they are not similar, the two are complements of each other. This was the best way found for now. -->
<!--  -->
<!-- Projects like in  https://github.com/ledongthuc/pdf and in https://github.com/mazeForGit/pdf were tried first instead of Pypdf2 but I didn't find better or equal results like in the python lib. -->

## Future Work

The development of a more thematic loading animation are in future plans, in addition to some future user recommendations.

<!-- ## Acknowledgments -->

<!-- Lumus was inspired by [project-name], [link-to-project]. -->

<!-- Special thanks to [contributor-names] for their contributions to the project. -->
