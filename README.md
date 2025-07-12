# 🕷️ CR4WL3R - Threat Intelligence Web Crawler

A sophisticated autonomous web crawler designed for threat intelligence gathering and dark web monitoring.

## 🎯 Features

- **🕸️ Autonomous Crawling**: Breadth-first search with configurable depth limits
- **🌐 Tor Integration**: Automatic .onion site detection and Tor routing
- **🔍 Threat Intelligence**: Content analysis and threat tagging
- **📊 Structured Output**: JSON-based results with metadata extraction
- **⚡ Async Performance**: High-performance asynchronous crawling
- **🛡️ Smart Filtering**: Domain-scoped crawling with external link discovery

## 🚀 Quick Start

### Prerequisites
```bash
# Python 3.8+
python --version

# Install dependencies
pip install -r threat_crawler/requirements.txt
```

### Basic Usage
```bash
cd threat_crawler
python main.py
```

### Configuration
Edit `threat_crawler/config/settings.py`:
```python
SEED_URL = "https://example.com"  # Your target URL
CRAWL_DEPTH_LIMIT = 3             # How deep to crawl
MAX_PAGES_PER_DOMAIN = 50         # Max pages per domain
```

## 📁 Project Structure

```
CR4WL3R/
├── threat_crawler/
│   ├── main.py              # Entry point
│   ├── config/
│   │   └── settings.py      # Configuration
│   ├── core/
│   │   ├── crawler.py       # Main crawling logic
│   │   ├── parser.py        # HTML parsing
│   │   ├── detector.py      # Site type detection
│   │   └── tagger.py        # Threat tagging
│   ├── fetcher/
│   │   └── client.py        # HTTP client (Tor support)
│   ├── storage/
│   │   └── writer.py        # JSON result storage
│   └── utils/
│       ├── link_utils.py    # Link extraction
│       └── logger.py        # Logging utilities
├── venv/                    # Virtual environment
├── .gitignore              # Git ignore rules
└── README.md               # This file
```

## 🔧 Configuration

### Tor Setup (Optional)
For .onion site crawling:
1. Install Tor Browser or Tor service
2. Ensure Tor is running on `127.0.0.1:9050`
3. The crawler automatically detects .onion URLs

### Target Types
- **Regular Sites**: `https://example.com`
- **Onion Sites**: `http://example.onion`
- **Mixed Crawling**: Configure multiple seed URLs

## 📊 Output Format

Results are saved to `output/results.json`:
```json
[
  {
    "url": "https://example.com",
    "status_code": 200,
    "type": "Forum",
    "title": "Example Forum",
    "tech_stack": ["WordPress", "JavaScript"],
    "headers": {},
    "tags": ["chat_forum", "exploit_market"]
  }
]
```

## 🛡️ Threat Intelligence

The crawler automatically tags content for:
- **Carding**: Credit card fraud indicators
- **Credentials**: Password/username dumps
- **Exploit Markets**: Malware/exploit sales
- **PII Leaks**: Personal information exposure
- **Chat Forums**: Underground communication platforms

## ⚠️ Legal & Ethical Considerations

- **Authorized Use Only**: Only crawl sites you own or have permission to crawl
- **Rate Limiting**: Respect robots.txt and implement delays
- **Data Privacy**: Handle sensitive data responsibly
- **Compliance**: Follow local laws and regulations

## 🔄 Development

### Adding New Threat Tags
Edit `threat_crawler/core/tagger.py`:
```python
def tag_content(html: str, headers: dict, tech_stack: list) -> list:
    tags = []
    # Add your custom detection logic
    if "your_keyword" in html.lower():
        tags.append("your_tag")
    return tags
```

### Extending Site Detection
Edit `threat_crawler/core/detector.py`:
```python
def detect_site_type(html):
    html_lower = html.lower()
    if "your_indicator" in html_lower:
        return "Your Site Type"
    return "Unknown"
```

## 📝 License

This project is for educational and authorized security research purposes only.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

**⚠️ Disclaimer**: This tool is designed for legitimate security research and threat intelligence gathering. Users are responsible for ensuring compliance with applicable laws and regulations. 