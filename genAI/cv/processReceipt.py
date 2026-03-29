import base64
import json
import requests
from io import BytesIO
from PIL import Image
from mindee import Client, product

class ProcessReceipts:
    def __init__(self, apiKey):
        self.apiKey = apiKey

    def convertImageToData(self, image: str, currency: str) -> dict:
        # Creating the image from the bytes sent by the backend
        imageBytes = base64.b64decode(image)
        pilImage = Image.open(BytesIO(imageBytes))

        # TODO: Write the logic to extract data from the receipt image using Tessaract OCV and send back to the backend
        
        # Send the response back to the backend
        receiptData: dict = {}
        return receiptData
