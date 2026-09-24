#!/bin/bash
sed -i 's/AddChunk(\(.*\))/AddChunk(\1, false)/g' wails-app/backend/classifier/buffer_test.go
