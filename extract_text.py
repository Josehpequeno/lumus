import PyPDF2
import sys 

def extract_text_from_pdf(pdf_path, page_number):
    text = ""
    with open(pdf_path, 'rb') as file:
        reader = PyPDF2.PdfReader(file)
        page = reader.pages[page_number -1] 
        text = page.extract_text()
    print(text)

pdf_path = sys.argv[1]
page_number = int(sys.argv[2])

extract_text_from_pdf(pdf_path, page_number)
