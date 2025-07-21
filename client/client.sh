#!/bin/bash

# A script to launch multiple client instances for testing,
# each served by its own Python HTTP server on a unique port.

# Validate input: Check if the number of clients is provided.
if [ -z "$1" ]; then
  echo "Usage: $0 <number_of_clients>"
  echo "Example: $0 3"
  exit 1
fi

NUM_CLIENTS=$1
BASE_PORT=3000
PIDS=() # Array to store the Process IDs of the background servers

# Function to clean up all background server processes on exit.
cleanup() {
  echo -e "\nShutting down all client servers..."
  for pid in "${PIDS[@]}"; do
    # Kill the process silently, ignoring errors if it's already gone.
    kill "$pid" 2>/dev/null
  done
  echo "Cleanup complete."
}

# Register the cleanup function to run when the script exits.
trap cleanup EXIT

# Loop to start a server for each client.
for (( i=0; i<$NUM_CLIENTS; i++ )); do
  PORT=$((BASE_PORT + i))
  URL="http://localhost:$PORT"

  echo "-> Starting server for client $((i+1)) on port $PORT..."
  # Start a Python server in the background to serve index.html from the current directory.
  python3 -m http.server "$PORT" &
  PIDS+=($!) # Store the PID of the last background process.

  echo "   Client $((i+1)) URL: $URL"

  # Open the URL in the default browser (cross-platform).
  [[ -x "$(command -v xdg-open)" ]] && xdg-open "$URL" >/dev/null 2>&1 || open "$URL" >/dev/null 2>&1 &
done

echo -e "\nAll $NUM_CLIENTS clients launched. Press Ctrl+C to shut down all servers."

# Wait indefinitely, allowing the user to terminate with Ctrl+C, which triggers the trap.
wait

