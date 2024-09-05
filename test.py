import requests
from requests.exceptions import RequestException, HTTPError, ConnectionError, Timeout

def check_http_port(port):
    url = f"http://localhost:{port}"
    try:
        response = requests.get(url)
        response.raise_for_status()  # Raise an HTTPError for bad responses (4xx and 5xx)
        if response.status_code == 200:
            print(f"Response from port {port}:")
            print(response.text)
        else:
            print(f"Received status code {response.status_code} from port {port}.")
    except Exception as e:
        print(f"Exception: {e}")

if __name__ == "__main__":
    while True:
        port = input("Enter the port number: ")
        try:
            port = int(port)
            check_http_port(port)
        except ValueError:
            print("Invalid port number. Please enter a valid integer.")
