import argparse
import requests
import re
import random
import string
import threading
import hashlib
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from urllib.parse import urlparse

requests.packages.urllib3.disable_warnings()

# List of known file extensions
KNOWN_EXTENSIONS = ['aspx', 'asp', 'cgi', 'pl', 'html', 'php', 'phtml']

def parse_args():
    parser = argparse.ArgumentParser(description='Fuzzing script to detect file extensions and average word count.')
    parser.add_argument('-u', '--url', required=True, help='URL to fuzz (e.g., http://example.com/test.aspx)')
    parser.add_argument('-ac', '--autocalibrate', type=int, default=3, help='Number of requests for autocalibration (default: 3)')
    parser.add_argument('-t', '--threads', type=int, default=5, help='Number of concurrent threads (default: 5)')
    parser.add_argument('-s', '--smart', action='store_true', help='Enable smart mode to check for specific server names')
    return parser.parse_args()

def is_file_name(url):
    """Check if the URL has a known file extension."""
    return any(url.endswith(f'.{ext}') for ext in KNOWN_EXTENSIONS)

def get_server_extension(url, smart_mode):
    """Get the server's file extension based on the response header."""
    try:
        response = requests.get(url, verify=False)
        server = response.headers.get('Server', '').lower()
        
        # List of server names that should trigger an exit
        blocking_servers = ["akamai", "cloudflare"]  # Add more as needed
        
        # Check for blocking servers in the server name
        if smart_mode and any(block in server for block in blocking_servers):
            print("[-] Detected Akamai or Cloudflare protection. Exiting...")
            exit(0)

        # Determine extension based on server type
        if 'apache' in server:
            return 'php'
        elif 'iis' in server:
            return 'aspx'
        elif 'nginx' in server:
            return 'html'
        else:
            return None
    except requests.RequestException as e:
        print(f"[-] Error fetching URL: {e}")
        return None

def generate_random_string(length=8):
    """Generate a random string of fixed length."""
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(length))

def get_word_count(base_url, autocalibrate):
    """Count words in the response text from random requests."""
    total_word_count = 0
    lock = threading.Lock()
    
    def make_request():
        nonlocal total_word_count
        # Generate a random URL based on whether it's a filename or directory
        if is_file_name(base_url):
            random_url = f"{base_url[:base_url.rfind('/') + 1]}{generate_random_string()}.{base_url.split('.')[-1]}"
        else:
            random_url = f"{base_url}/{generate_random_string()}"

        try:
            response = requests.get(random_url, verify=False)
            word_count = len(re.findall(r'\w+', response.text))
            with lock:
                total_word_count += word_count
        except requests.RequestException as e:
            pass

    threads_list = []
    
    for _ in range(autocalibrate):
        thread = threading.Thread(target=make_request)
        threads_list.append(thread)
        thread.start()

    for thread in threads_list:
        thread.join()

    average_word_count = total_word_count / autocalibrate if autocalibrate > 0 else 0
    return average_word_count

def calculate_md5(url):
    """Calculate MD5 hash of the given URL."""
    return hashlib.md5(url.encode()).hexdigest()

def main():
    args = parse_args()
    
    url = args.url
    extension = None
    
    # Check if the URL is a filename with an extension or a directory
    if is_file_name(url):
        extension = url.split('.')[-1]  # Get the extension from the URL
        fuzzed_url = f"{url[:url.rfind('/') + 1]}FUZZ"  # Replace filename with FUZZ
    else:
        # If not, fetch from server and use original URL path for FUZZ
        extension = get_server_extension(url, args.smart)
        
        # If no extension is found, default to common extensions: php and asp.
        if not extension:
            #print("[-] Could not determine file extension. Defaulting to common extensions: php and asp.")
            extensions_to_test = ['php', 'asp']
            
            # Prepare fuzzed URLs for each extension to test
            for ext in extensions_to_test:
                fuzzed_url_php = f"{url}/FUZZ.{ext}"
                #print(f"Testing: {fuzzed_url_php}")

            # Print ffuf command with first extension and others as -e flag
            first_extension = extensions_to_test[0]
            additional_extensions = ','.join(extensions_to_test[1:])
            
            print(f"ffuf -ac -fc 403 -w lower_without_dots -u {url}/FUZZ.{first_extension} -e {additional_extensions}")
            return
        
        fuzzed_url = f"{url}/FUZZ"

    # Make a request to get word count with autocalibration
    average_word_count = get_word_count(url, args.autocalibrate)
    
    # Calculate MD5 hash of the supplied URL
    md5_hash = calculate_md5(url)

    print(f"ffuf -ac -fc 403 -w lower_without_dots -u {fuzzed_url}.{extension} -fw {average_word_count:.0f} -o {md5_hash}")

if __name__ == '__main__':
    main()
