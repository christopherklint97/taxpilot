#!/usr/bin/env bash
# build-knowledge.sh — Build and persist the knowledge base
#
# This script is a stub for the knowledge base build pipeline. When the
# full content extraction pipeline is in place, this script will:
#
# 1. Run extract-instructions.sh to get raw document chunks
# 2. Optionally generate embeddings for vector search (future)
# 3. Persist the knowledge store to data/knowledge/
# 4. Validate the store by running search sanity checks
#
# For the MVP, the knowledge base uses TF-IDF keyword search with
# hand-curated seed documents (no embeddings needed). A future upgrade
# path would be:
#
#   a. Generate embeddings using an API (OpenAI, Voyage, or local model)
#   b. Store embeddings in SQLite with the vec extension
#   c. Switch the Store.Search() method to use cosine similarity
#
# Prerequisites:
#   - Go 1.22+
#   - (Future) Embedding API key or local embedding model
#
# Usage:
#   ./scripts/build-knowledge.sh [--year 2025] [--with-embeddings]

set -euo pipefail

YEAR="${1:-2025}"
DATA_DIR="data/knowledge"

echo "build-knowledge.sh — Knowledge base builder (stub)"
echo ""
echo "This script is not yet implemented. The knowledge base is currently"
echo "seeded at runtime from internal/knowledge/seed.go."
echo ""
echo "Future pipeline:"
echo "  1. Extract text from IRS/FTB PDFs (extract-instructions.sh)"
echo "  2. Generate embeddings (optional, for vector search)"
echo "  3. Persist to ${DATA_DIR}/"
echo "  4. Load at runtime via knowledge.NewStoreFromDir()"
echo ""
echo "Target tax year: ${YEAR}"
echo "Output directory: ${DATA_DIR}/"
