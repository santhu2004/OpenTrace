# core/crawler.py

import asyncio
import json
import subprocess
import tempfile
import os
import sys
from pathlib import Path
from typing import Dict, List, Any
from storage.writer import save_result
from config.settings import CRAWL_DEPTH_LIMIT, MAX_PAGES_PER_DOMAIN


def run_go_crawler(config: dict) -> List[Dict[str, Any]]:
    """
    Runs the Go-based fastcrawl.exe with the given config, captures its JSON output, and returns the results.
    Uses subprocess.Popen to read output line by line and avoid deadlocks.
    Args:
        config: Dictionary with keys: start_url, max_depth, max_pages, timeout, user_agent, workers
    Returns:
        List of parsed JSON results from the Go crawler
    Raises:
        RuntimeError if the Go binary fails or output is invalid
    """
    import time
    import threading
    import queue
    import os
    import json

    go_bin = os.path.join(os.path.dirname(__file__), '../go_crawler/fastcrawl.exe')
    if not os.path.exists(go_bin):
        go_bin = os.path.join(os.path.dirname(__file__), '../go_crawler/fastcrawl')
        if not os.path.exists(go_bin):
            raise FileNotFoundError(f"Go crawler binary not found at {go_bin}")

    go_dir = os.path.dirname(go_bin)

    required = ['start_url', 'max_depth', 'max_pages', 'timeout', 'user_agent', 'workers']
    for key in required:
        if key not in config or config[key] in (None, ""):
            raise ValueError(f"Missing required config: {key}")

    command = [
        go_bin,
        f"-start_url={config['start_url']}",
        f"-max_depth={int(config['max_depth'])}",
        f"-max_pages={int(config['max_pages'])}",
        f"-timeout={str(config['timeout'])}",
        f"-user_agent={str(config['user_agent'])}",
        f"-workers={int(config['workers'])}",
    ]
    print(f"[DEBUG] Running Go crawler command: {' '.join(command)}")
    print(f"[DEBUG] Working directory for Go: {go_dir}")

    results = []
    start_time = time.time()
    try:
        proc = subprocess.Popen(
            command,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding="utf-8",
            bufsize=1,
            cwd=go_dir,
        )
        print(f"[DEBUG] Go crawler PID: {proc.pid}")

        # Print Go stderr in a background thread
        def stderr_reader():
            for line in proc.stderr:
                print(f"[GO-STDERR] {line.rstrip()}")
        stderr_thread = threading.Thread(target=stderr_reader, daemon=True)
        stderr_thread.start()

        # Watchdog for output timeout
        q = queue.Queue(maxsize=1000)
        def reader():
            for line in proc.stdout:
                line = line.strip()
                if not line:
                    continue
                print(f"[GO-OUT] {line}")
                try:
                    result = json.loads(line)
                    q.put(result)
                except json.JSONDecodeError:
                    print(f"[WARN] Could not decode JSON: {line}")
                    continue
            print("[DEBUG] Go stdout closed.")
        t = threading.Thread(target=reader, daemon=True)
        t.start()
        last_output = time.time()
        while True:
            try:
                result = q.get(timeout=10)
                last_output = time.time()
                results.append(result)
            except queue.Empty:
                if proc.poll() is not None and q.empty():
                    print(f"[DEBUG] Go process exited with code {proc.returncode} and output queue is empty.")
                    break
                elif proc.poll() is None:
                    print("[WARN] No JSON output from Go for 10s, but Go is still running...")
                    continue
                else:
                    print(f"[DEBUG] Go process exited with code {proc.returncode} but output queue not empty.")
                    break
        proc.stdout.close()
        proc.wait()
        print(f"[DEBUG] Go process final exit code: {proc.returncode}")
        if proc.returncode != 0:
            print(f"[ERROR] Go crawler failed: return code {proc.returncode}")
            # Print any remaining stderr
            try:
                err = proc.stderr.read()
                if err:
                    print(f"[GO-STDERR-FINAL] {err}")
            except Exception:
                pass
            raise RuntimeError(f"Go crawler failed: return code {proc.returncode}")
        if not results:
            print("[ERROR] Go crawler produced no output.")
            raise RuntimeError("Go crawler produced no output.")
    except Exception as e:
        print(f"[ERROR] Exception while running Go crawler: {e}")
        raise
    print(f"[DEBUG] Parsed {len(results)} results from Go crawler in {time.time() - start_time:.2f}s.")
    from storage.writer import save_result
    save_result(results)
    print(f"[✓] Results saved to output/results.ndjson ({len(results)} entries)")
    print(f"[INFO] Total links crawled: {len(results)}")
    return results

def convert_go_results_to_python_format(go_output: Dict[str, Any]) -> List[Dict[str, Any]]:
    """
    Convert Go crawler results to Python pipeline format.
    
    Args:
        go_output: Raw output from Go crawler
        
    Returns:
        List of page results in Python format
    """
    python_results = []
    
    for result in go_output['results']:
        # Convert Go result to Python format
        python_result = {
            "url": result['url'],
            "status_code": result['status'],
            "title": result['title'],
            "headers": result['headers'],
            "links": result['links'],
            "internal_links": result['internal_links'],
            "external_links": result['external_links'],
            "depth": result['depth'],
            "discovered": result['discovered']
        }
        
        # Add error if present
        if 'error' in result and result['error']:
            python_result['error'] = result['error']
        
        python_results.append(python_result)
    
    return python_results

async def crawl_site(seed_url):
    """
    Autonomous scoped crawler that explores links within the same domain.
    Now uses Go crawler for high-performance crawling.
    """
    print(f"[+] Starting Go-powered crawl for: {seed_url}")
    
    try:
        # Use Go crawler for high-performance crawling
        go_results = run_go_crawler(
            target_url=seed_url,
            max_depth=CRAWL_DEPTH_LIMIT,
            max_links=MAX_PAGES_PER_DOMAIN
        )
        
        # Convert to Python format
        pages = convert_go_results_to_python_format(go_results)
        
        print(f"[+] Processing {len(pages)} crawled pages with Python threat intelligence...")
        
        # Process each page with Python threat intelligence modules
        for page in pages:
            if page['status_code'] != 200:
                print(f"[!] Skipping failed page: {page['url']} (status: {page['status_code']})")
                continue

            # For now, we'll use the data from Go crawler
            # In a full implementation, you might want to re-fetch with Python
            # to get HTML for detailed analysis
            print(f"[+] Processing: {page['url']}")
            
            # Create a basic result structure compatible with existing pipeline
            result_data = {
                "url": page['url'],
                "status_code": page['status_code'],
                "title": page['title'],
                "headers": page['headers'],
                "links": page['links'],
                "internal_links": page['internal_links'],
                "external_links": page['external_links'],
                "depth": page['depth'],
                "discovered": page['discovered'],
                # Placeholder for Python analysis results
                "type": "unknown",  # Will be filled by detector
                "tech_stack": [],   # Will be filled by parser
                "tags": []          # Will be filled by tagger
            }
            
            # Save the result
            await save_result(result_data)
        
        print(f"[✓] Completed Go-powered crawl for {seed_url}")
        print(f"[*] Summary: {go_results['summary']['total_pages']} pages, "
              f"{go_results['summary']['successful']} successful, "
              f"{go_results['summary']['failed']} failed")
        
    except Exception as e:
        print(f"[!] Go crawler integration failed: {e}")
        raise
