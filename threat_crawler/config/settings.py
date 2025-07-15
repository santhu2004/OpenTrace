# config/settings.py — Constants for timeout, headers, etc.

# # Add clearnet and .onion targets
# TARGETS = [
#     "http://example.com",                         # Standard test page (safe)
#     "https://httpbin.org/html",                   # Simple HTML page for testing
#     "https://news.ycombinator.com",               # Forum-style site (YCombinator)
#     "https://www.reddit.com/r/cybersecurity/",    # Subreddit page (complex DOM)
#     "https://pastebin.com",                       # Paste site (some content gated)
#     "https://dumps.pw",                           # Known credential leak aggregator (NSFW)
#     "https://raidforums.wiki",                    # Archive of RaidForums info (non-malicious)
#     "https://intelx.io",                          # Threat Intel search engine (may block crawling)
#     "https://0paste.com",                         # Paste site (can be slow)
#     "http://hqfld5smkr4b4xrjcco7zotvoqhuuoehjdvoin755iytmpk4sm7cbwad.onion/",
#     "http://2ln3x7ru6psileh7il7jot2ufhol4o7nd54z663xonnnmmku4dgkx3ad.onion"   
# ]
# # config/settings.py

# Example .onion sites (uncomment to test Tor functionality):
# SEED_URL = "http://hqfld5smkr4b4xrjcco7zotvoqhuuoehjdvoin755iytmpk4sm7cbwad.onion/"
# SEED_URL = "http://2ln3x7ru6psileh7il7jot2ufhol4o7nd54z663xonnnmmku4dgkx3ad.onion"



HEADERS = {
    "User-Agent": "Mozilla/5.0 (ThreatIntelBot)"
}

TIMEOUT = '10s'  # Go expects string like '10s'
# Tor is now automatically enabled for .onion sites
TOR_PROXY = "socks5://127.0.0.1:9050"

# Standardized config for Go integration
START_URL = "https://hellofhackers.com"
MAX_DEPTH = 3
MAX_PAGES = 8000 #make it 8k later
TIMEOUT = '10s'  # Go expects string like '10s'
USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
WORKERS = 10
CRAWL_DEPTH_LIMIT = MAX_DEPTH
MAX_PAGES_PER_DOMAIN = MAX_PAGES
