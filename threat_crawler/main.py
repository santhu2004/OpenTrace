# main.py — Entry point for the hybrid Go/Python crawler
import asyncio
import os
import json
from pathlib import Path

from core.crawler import crawl_site
from config.settings import SEED_URL

async def main():
    """
    Main orchestration function that runs the hybrid Go/Python crawler.
    """
    print("🚀 Starting CR4WL3R - Hybrid Go/Python Threat Intelligence Crawler")
    print(f"🎯 Target: {SEED_URL}")
    print("=" * 60)
    
    # Ensure output directory exists
    os.makedirs("output", exist_ok=True)
    
    # Initialize results file
    results_file = Path("output/results.json")
    if results_file.exists():
        results_file.unlink()  # Remove existing file
    
    # Start with JSON array
    with open(results_file, "w", encoding="utf-8") as f:
        f.write("[\n")
    
    try:
        # Run the Go-powered crawler
        await crawl_site(SEED_URL)
        
        # Close the JSON array
        with open(results_file, "a", encoding="utf-8") as f:
            f.write("\n]\n")
        
        print("=" * 60)
        print("✅ Crawling completed successfully!")
        print(f"📁 Results saved to: {results_file}")
        
        # Show file size
        if results_file.exists():
            size_kb = results_file.stat().st_size / 1024
            print(f"📊 Results file size: {size_kb:.1f} KB")
        
    except Exception as e:
        print(f"❌ Crawling failed: {e}")
        raise
    finally:
        # Ensure JSON array is properly closed even on error
        if results_file.exists() and results_file.stat().st_size > 0:
            with open(results_file, "r", encoding="utf-8") as f:
                content = f.read()
            
            if not content.strip().endswith("]"):
                with open(results_file, "a", encoding="utf-8") as f:
                    f.write("\n]\n")

if __name__ == "__main__":
    asyncio.run(main())
