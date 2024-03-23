import PyPDF2
import sys 
import re

def extract_text_from_pdf(pdf_path, page_number):
    text = ""
    with open(pdf_path, 'rb') as file:
        reader = PyPDF2.PdfReader(file)
        page = reader.pages[page_number -1] 
        text = page.extract_text()
        size = page.cropbox.upper_right
        width, height =  size

        print(f"{width}")
        print(f"{height}")
        # Remove excess spaces
        # text_without_excessive_spaces = re.sub(r'\s{1,6}', ' ', text).strip()
        text_without_excessive_spaces = re.sub(r'\s+', ' ', text).strip()
        print(text_without_excessive_spaces)

pdf_path = sys.argv[1]
page_number = int(sys.argv[2])

extract_text_from_pdf(pdf_path, page_number)
