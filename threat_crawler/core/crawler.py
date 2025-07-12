# core/crawler.py

import asyncio
import json
import subprocess
import tempfile
import os
from pathlib import Path
from typing import Dict, List, Any
from storage.writer import save_result
from config.settings import CRAWL_DEPTH_LIMIT, MAX_PAGES_PER_DOMAIN

def run_go_crawler(target_url: str, max_depth: int = 2, max_links: int = 50) -> Dict[str, Any]:
    """
    Run the Go fastcrawl binary as a subprocess and return structured results.
    
    Args:
        target_url: URL to crawl
        max_depth: Maximum crawl depth
        max_links: Maximum number of pages to crawl
        
    Returns:
        Dictionary containing crawl results and summary
    """
    # Path to the Go binary
    go_binary_path = Path(__file__).parent.parent / "go_crawler" / "fastcrawl.exe"
    
    if not go_binary_path.exists():
        raise FileNotFoundError(f"Go crawler binary not found: {go_binary_path}")
    
    # Prepare configuration for Go crawler
    config = {
        "target_url": target_url,
        "max_depth": max_depth,
        "max_links": max_links,
        "timeout": "30s",
        "max_concurrency": 10,
        "user_agent": "ThreatCrawler/1.0",
        "respect_robots": False
    }
    
    try:
        # Create temporary input file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as temp_input:
            json.dump(config, temp_input)
            temp_input_path = temp_input.name
        
        # Prepare command
        cmd = [str(go_binary_path), "-input", temp_input_path]
        
        # Execute Go crawler
        print(f"[+] Running Go crawler for: {target_url}")
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=300  # 5 minute timeout
        )
        
        # Handle errors
        if result.returncode != 0:
            print(f"[!] Go crawler failed with exit code {result.returncode}")
            print(f"[!] Error output: {result.stderr}")
            raise subprocess.CalledProcessError(result.returncode, cmd, result.stdout, result.stderr)
        
        # Parse JSON output
        output_data = json.loads(result.stdout)
        
        # Validate output structure
        required_keys = ['config', 'results', 'summary', 'timestamp']
        for key in required_keys:
            if key not in output_data:
                raise ValueError(f"Missing required key in Go crawler output: {key}")
        
        print(f"[✓] Go crawler completed successfully")
        print(f"[*] Crawled {output_data['summary']['total_pages']} pages")
        print(f"[*] Duration: {output_data['summary']['duration']}")
        
        return output_data
        
    except json.JSONDecodeError as e:
        print(f"[!] Failed to parse Go crawler output: {e}")
        raise
    except subprocess.TimeoutExpired:
        print(f"[!] Go crawler timed out after 5 minutes")
        raise
    finally:
        # Clean up temporary file
        if 'temp_input_path' in locals():
            os.unlink(temp_input_path)

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
