# CR4WL3R - Hybrid Go/Python Threat Intelligence Crawler

A high-performance web crawler that combines the speed of Go with the flexibility of Python for threat intelligence gathering.

## 🚀 Features

- **High Performance**: Go-based concurrent crawling
- **Threat Intelligence**: Python-based analysis pipeline
- **Hybrid Architecture**: Best of both languages
- **Configurable**: Easy to customize crawl parameters
- **Production Ready**: Clean, minimal codebase

## 📁 Project Structure

```
threat_crawler/
├── main.py                 # Entry point
├── core/
│   └── crawler.py         # Go integration & orchestration
├── config/
│   └── settings.py        # Configuration
├── storage/
│   └── writer.py          # Result storage
├── go_crawler/
│   ├── fastcrawl.exe      # Compiled Go binary
│   ├── *.go              # Go source files
│   ├── go.mod            # Go module
│   └── README.md         # Go module documentation
├── output/                # Crawl results
├── requirements.txt       # Python dependencies
├── HYBRID_ARCHITECTURE.md # Architecture documentation
└── README.md             # This file
```

## 🛠️ Installation

1. **Clone the repository**
2. **Build the Go binary**:
   ```bash
   cd go_crawler
   go build -o fastcrawl.exe
   ```
3. **Install Python dependencies** (minimal):
   ```bash
   pip install -r requirements.txt
   ```

## 🎯 Usage

### Basic Usage
```bash
python main.py
```

### Configuration
Edit `config/settings.py` to customize:
- Target URL
- Crawl depth
- Maximum pages
- Timeouts

### Output
Results are saved to `output/results.json` in JSON format.

## 🔧 How It Works

1. **Python** orchestrates the crawl and manages configuration
2. **Go binary** performs high-speed concurrent crawling
3. **Python** processes results and applies threat intelligence
4. **Results** are saved to JSON files

## 📊 Performance

- **Concurrent crawling** with configurable worker count
- **Fast link extraction** and processing
- **Efficient memory usage**
- **Timeout protection**

## 🎯 Use Cases

- **Threat Intelligence**: Gather data from suspicious sites
- **Security Research**: Analyze web infrastructure
- **Content Discovery**: Find related pages and resources
- **Link Analysis**: Map site structure and relationships

## 📝 License

This project is for educational and research purposes.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

**Built with ❤️ for the security community** 