# main.py — Entry point for the hybrid Go/Python crawler
import sys
import traceback
from config import settings
from core.crawler import run_go_crawler
from core.tagger import tag_content
from core.parser import extract_page_info
from storage.writer import save_result


def main():
    """
    Main entry point for CR4WL3R Python pipeline.
    Loads config, runs Go crawler, analyzes, and saves results.
    """
    # Load config from settings.py
    config = {
        'start_url': getattr(settings, 'START_URL', None),
        'max_depth': getattr(settings, 'MAX_DEPTH', 3),
        'max_pages': getattr(settings, 'MAX_PAGES', 100),
        'timeout': getattr(settings, 'TIMEOUT', '30s'),
        'user_agent': getattr(settings, 'USER_AGENT', 'ThreatCrawler/3.0'),
        'workers': getattr(settings, 'WORKERS', 10),
    }
    print(f"[DEBUG] Using config: {config}")

    try:
        # Run Go crawler and get results
        go_results = run_go_crawler(config)
        print(f"[DEBUG] Go crawler returned {len(go_results)} results.")
    except Exception as e:
        print(f"[ERROR] Failed to run Go crawler: {e}", file=sys.stderr)
        traceback.print_exc()
        sys.exit(1)

    if not go_results:
        print("[WARN] Go crawler returned no results. No output will be saved.")

    processed_results = []
    for result in go_results:
        # Optionally analyze with parser/tagger
        html = result.get('html', '')  # If HTML is included in Go output
        title, tech_stack = extract_page_info(html) if html else (result.get('title', ''), [])
        tags = tag_content(html, result.get('headers', {}), tech_stack)
        result['title'] = title
        result['tech_stack'] = tech_stack
        result['tags'] = tags
        processed_results.append(result)

    # Save results to output/results.json
    try:
        print(f"[DEBUG] Saving {len(processed_results)} results to output/results.json")
        save_result(processed_results)
        print(f"[✓] Results saved to output/results.json")
    except Exception as e:
        print(f"[ERROR] Failed to save results: {e}", file=sys.stderr)
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()
