import json
import os

def save_result(results):
    """
    Save a list of results to output/results.ndjson in NDJSON format (one JSON object per line).
    Overwrites any existing file.
    """
    OUTPUT_FILE = "output/results.ndjson"
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    abs_path = os.path.abspath(OUTPUT_FILE)
    try:
        with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
            for result in results:
                f.write(json.dumps(result, ensure_ascii=False) + "\n")
        print(f"[✓] Results saved to {abs_path} (NDJSON format)")
    except Exception as e:
        print(f"[ERROR] Failed to write NDJSON results file: {e}")
        raise

def save_tagged_result(results):
    """
    Save a list of tagged results to output/tagged.ndjson in NDJSON format (one JSON object per line).
    Overwrites any existing file.
    """
    OUTPUT_FILE = "output/tagged.ndjson"
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    abs_path = os.path.abspath(OUTPUT_FILE)
    try:
        with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
            for result in results:
                f.write(json.dumps(result, ensure_ascii=False) + "\n")
        print(f"[✓] Tagged results saved to {abs_path} (NDJSON format)")
    except Exception as e:
        print(f"[ERROR] Failed to write tagged NDJSON results file: {e}")
        raise
