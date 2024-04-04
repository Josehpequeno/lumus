package main

import (
	"fmt"
	"log"

	"code.sajari.com/docconv/v2"
)

func main() {
	// Caminho do arquivo PDF
	pdfPath := "../example2.pdf"

	// Página específica que você deseja extrair (por exemplo, página 1)
	pageNumber := 1

	// Extraindo texto da página específica do PDF
	text, err := extractTextFromPage(pdfPath, pageNumber)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Texto extraído da página", pageNumber, ":", text)
}

func extractTextFromPage(pdfPath string, pageNumber int) (string, error) {
	// Convertendo o arquivo PDF para texto com docconv
	res, err := docconv.ConvertPath(pdfPath)
	if err != nil {
		return "", err
	}

	// Verificando se o número da página é válido
	if pageNumber < 1 || pageNumber > len(res.Pages) {
		return "", fmt.Errorf("página %d não encontrada", pageNumber)
	}

	// Retornando o texto da página especificada
	return res.Pages[pageNumber-1].Text, nil
}
