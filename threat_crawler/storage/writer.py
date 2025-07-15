import json
import os

def save_result(results):
    """
    Save a list of results to output/results.json as a JSON array.
    Overwrites any existing file. Asserts file exists and is valid JSON after writing.
    """
    OUTPUT_FILE = "output/results.json"
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    abs_path = os.path.abspath(OUTPUT_FILE)
    try:
        with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
            json.dump(results, f, ensure_ascii=False, indent=2)
        # Verify file exists and is valid JSON
        assert os.path.exists(OUTPUT_FILE), f"[ERROR] Output file {abs_path} was not created!"
        with open(OUTPUT_FILE, "r", encoding="utf-8") as f:
            loaded = json.load(f)
        assert isinstance(loaded, list), f"[ERROR] Output file {abs_path} does not contain a JSON array!"
        print(f"[✓] Results saved to {abs_path}")
    except Exception as e:
        print(f"[ERROR] Failed to write or verify results file: {e}")
        raise
