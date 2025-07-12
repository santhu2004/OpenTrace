import json
import os
import asyncio

# Optional: Thread-safe write lock if you plan to write from multiple coroutines
lock = asyncio.Lock()
OUTPUT_FILE = "output/results.json"

# Ensure output directory exists
os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)

# Track if we've written any results
results_written = False

async def save_result(result):
    global results_written

    async with lock:
        # Initialize file with opening bracket if this is the first write
        if not results_written:
            with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
                f.write("[\n")
        
        # Append the result
        with open(OUTPUT_FILE, "a", encoding="utf-8") as f:
            if results_written:
                f.write(",\n")  # Add comma before new object if not the first
            json.dump(result, f, ensure_ascii=False, indent=2)
            results_written = True
