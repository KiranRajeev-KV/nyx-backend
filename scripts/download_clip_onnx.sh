#!/bin/bash

# Download CLIP ViT-B/32 ONNX model for local image embeddings
# Skips download if model already exists

MODEL_DIR="models/clip"
mkdir -p "$MODEL_DIR"

# Check if already downloaded
if [ -f "$MODEL_DIR/vision_model.onnx" ]; then
    echo "CLIP ONNX model already exists at $MODEL_DIR/vision_model.onnx"
    echo "To re-download, delete the file manually: rm $MODEL_DIR/vision_model.onnx"
    exit 0
fi

echo "Downloading CLIP ViT-B/32 ONNX model..."
echo "This may take a few minutes depending on your connection..."

# Try downloading from HuggingFace
# Using the microsoft CLIP model which has ONNX export
wget -O "$MODEL_DIR/vision_model.onnx" \
    "https://huggingface.co/microsoft/clip-vit-base-32/resolve/main/onnx/model.onnx" \
    --progress=bar:force:noscroll \
    2>&1

if [ $? -eq 0 ]; then
    echo "Download complete!"
    echo "Model saved to: $MODEL_DIR/vision_model.onnx"
    
    # Check file size
    SIZE=$(du -h "$MODEL_DIR/vision_model.onnx" | cut -f1)
    echo "Model size: $SIZE"
else
    echo "wget failed. Trying alternative method..."
    
    # Alternative: Use huggingface-cli or python
    echo "Please install huggingface-hub and run:"
    echo "  huggingface-cli download microsoft/clip-vit-base-32 --include onnx/model.onnx --local models/clip/"
    exit 1
fi
