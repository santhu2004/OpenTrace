import re
from langdetect import detect, LangDetectException
from collections import defaultdict
from bs4 import BeautifulSoup
import logging
import yaml
import os

# --- Broader keyword/regex lists ---
MARKETPLACE_KEYWORDS = [
    "market", "vendor", "listing", "escrow", "feedback", "buyer", "seller", "cart", "checkout", "order",
    "shipping", "product", "service", "purchase", "shop", "store", "deal", "review", "rating", "guarantee",
    "dispute", "wallet", "payment", "commission", "affiliate", "trusted seller", "buyer protection",
    # Expanded
    "vendor profile", "trusted vendor", "marketplace", "order history", "my orders", "my cart", "add to cart",
    "shipping address", "track order", "order status", "pending order", "completed order", "refund", "buyer review",
    "seller rating", "product listing", "featured vendor", "escrow service", "escrow enabled", "vendor bond",
    "vendor application", "vendor fee", "vendor panel", "vendor dashboard", "vendor support", "buyer support",
    "purchase protection", "buyer feedback", "vendor feedback", "market rules", "market admin", "market support"
]
MARKETPLACE_REGEX = [
    r"vendor(s)?\\s+list", r"escrow\\s+(service|enabled)", r"add\\s+to\\s+cart", r"leave\\s+feedback"
]

FORUM_KEYWORDS = [
    "forum", "thread", "post", "reply", "topic", "board", "discussion", "subforum", "member", "register",
    "join", "community", "message", "quote", "bump", "sticky", "announcement", "moderator", "admin",
    "user profile", "signature", "private message", "inbox", "outbox", "post count", "view posts", "edit post",
    "forum rules", "forum admin", "forum moderator", "forum staff", "forum index", "forum search", "forum statistics",
    "forum activity", "forum login", "forum register", "forum user", "forum member", "forum post", "forum reply",
    "forum thread", "forum topic", "forum announcement", "forum sticky", "forum bump", "forum quote", "forum message"
]
FORUM_REGEX = [
    r"new\\s+thread", r"reply\\s+to\\s+post", r"view\\s+topic", r"forum\\s+index"
]

PASTE_KEYWORDS = [
    "paste", "pastebin", "dump", "raw paste", "snippet", "share code", "upload text", "public paste",
    "private paste", "expiration", "syntax highlight", "clone", "fork", "raw", "bin", "hastebin", "ghostbin",
    # Expanded
    "new paste", "public pastes", "my pastes", "recent pastes", "paste url", "paste title", "paste content",
    "paste description", "paste password", "paste expire", "paste size", "paste language", "paste author",
    "paste created", "paste updated", "paste view", "paste download", "paste report", "paste share",
    "paste delete", "paste edit", "paste raw", "paste embed", "paste search", "paste trending", "paste api",
    "pastebin api", "pastebin pro", "pastebin login", "pastebin signup", "pastebin user", "pastebin guest"
]
PASTE_REGEX = [
    r"paste\\s+id", r"raw\\s+dump", r"view\\s+paste", r"create\\s+paste"
]

logger = logging.getLogger("tagger")
logging.basicConfig(level=logging.INFO)

# Load blocklist patterns at module level
BLOCKLIST_PATH = os.path.join(os.path.dirname(__file__), 'url_blocklist.yaml')
try:
    with open(BLOCKLIST_PATH, 'r', encoding='utf-8') as f:
        blocklist_patterns = yaml.safe_load(f)['patterns']
    blocklist_regex = [re.compile(p) for p in blocklist_patterns]
except Exception as e:
    blocklist_regex = []
    logger.warning(f"Could not load blocklist: {e}")

def is_blocked_url(url):
    return any(r.search(url) for r in blocklist_regex)

def extract_visible_text(html):
    soup = BeautifulSoup(html, "html.parser")
    for script in soup(["script", "style", "noscript"]):
        script.extract()
    text = soup.get_text(separator=" ")
    return " ".join(text.split()).lower()

def match_any(text, keywords, regexes):
    matches = []
    for kw in keywords:
        if kw in text:
            matches.append(kw)
    for pattern in regexes:
        if re.search(pattern, text):
            matches.append(pattern)
    return matches

class Tagger:
    def __init__(self):
        self.summary = defaultdict(int)

    def tag(self, record):
        html = record.get("html", "")
        url = record.get("url", "")
        # Blocklist check
        if is_blocked_url(url):
            logger.info(f"Blocked by URL blocklist: {url}")
            self.summary['blocked'] += 1
            return []
        text = extract_visible_text(html)
        tags = []
        match_info = {}

        # Marketplace
        mkt_matches = match_any(text, MARKETPLACE_KEYWORDS, MARKETPLACE_REGEX)
        if mkt_matches:
            tags.append("marketplace")
            match_info["marketplace"] = mkt_matches

        # Forum
        forum_matches = match_any(text, FORUM_KEYWORDS, FORUM_REGEX)
        if forum_matches:
            tags.append("forum")
            match_info["forum"] = forum_matches

        # Paste site
        paste_matches = match_any(text, PASTE_KEYWORDS, PASTE_REGEX)
        if paste_matches:
            tags.append("paste_site")
            match_info["paste_site"] = paste_matches

        # Darkweb
        if ".onion" in url:
            tags.append("darkweb")
            match_info["darkweb"] = [url]

        # Login page
        if re.search(r'<input[^>]+type=["\']?password["\']?', html, re.I):
            tags.append("login_page")
            match_info["login_page"] = ["password input"]

        # Marketplace/Phishing/Contact/Other (legacy logic)
        # ... (keep or expand as needed) ...

        # Language tag
        try:
            lang = detect(text)
            tags.append(f'lang_{lang}')
            match_info['lang'] = lang
        except LangDetectException:
            pass

        # Suspicious behavior
        redirects = record.get('redirects', 0)
        external_links = record.get('external_links', [])
        if isinstance(redirects, int) and redirects > 5:
            tags.append('suspicious_behavior')
            match_info['suspicious_behavior'] = f"redirects={redirects}"
        if isinstance(external_links, list) and len(external_links) > 20:
            tags.append('suspicious_behavior')
            match_info['suspicious_behavior'] = f"external_links={len(external_links)}"

        # Fallback: Uncategorized
        if not tags:
            tags.append("Uncategorized")
            match_info["uncategorized"] = True
            logger.warning(f"Uncategorized page: {url}")

        # Log tagging decision
        logger.info(f"URL: {url} | Tags: {tags} | Matches: {match_info}")
        for t in tags:
            self.summary[t] += 1
        return tags

def tag_content(html, headers, tech_stack):
    record = {'html': html, 'headers': headers, 'tech_stack': tech_stack}
    return Tagger().tag(record)

def tag_ndjson_records(records):
    tagger = Tagger()
    for rec in records:
        rec['tags'] = tagger.tag(rec)
    print("[Tagger] Tagging summary:")
    for tag, count in tagger.summary.items():
        print(f"  {tag}: {count}")
    return records 