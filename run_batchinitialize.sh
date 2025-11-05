#!/bin/bash

# Batch Initialize Participants Script
# Usage: ./run_batchinitialize.sh --deployment <dev|test|prod> --csv <path-to-csv>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Parse arguments
DEPLOYMENT=""
CSV_FILE=""
ENV_ROOT=""

while [[ $# -gt 0 ]]; do
	case $1 in
	--deployment)
		DEPLOYMENT="$2"
		shift 2
		;;
	--csv)
		CSV_FILE="$2"
		shift 2
		;;
	--env-root)
		ENV_ROOT="$2"
		shift 2
		;;
	*)
		echo "Unknown option: $1"
		echo "Usage: $0 --deployment <dev|test|prod> --csv <path-to-csv> [--env-root <dir>]"
		exit 1
		;;
	esac
done

# Validate required arguments
if [ -z "$DEPLOYMENT" ]; then
	echo "Error: --deployment is required (dev|test|prod)"
	exit 1
fi

if [ -z "$CSV_FILE" ]; then
	echo "Error: --csv is required"
	exit 1
fi

# Check if CSV file exists
if [ ! -f "$CSV_FILE" ]; then
	echo "Error: CSV file not found: $CSV_FILE"
	exit 1
fi

# Set env root to current directory if not provided
if [ -z "$ENV_ROOT" ]; then
	ENV_ROOT="."
fi

echo "=========================================="
echo "Batch Initialize Participants"
echo "=========================================="
echo "Deployment: $DEPLOYMENT"
echo "CSV File: $CSV_FILE"
echo "Env Root: $ENV_ROOT"
echo "=========================================="

# Build the batch initialize binary
echo "Building batch initialize binary..."
go build -o ./bin/batchinitialize ./cmd/batchinitialize

# Run the batch initialize
echo "Starting batch initialization..."
./bin/batchinitialize \
	--deployment "$DEPLOYMENT" \
	--csv "$CSV_FILE" \
	--env-root "$ENV_ROOT"

echo "Done!"
