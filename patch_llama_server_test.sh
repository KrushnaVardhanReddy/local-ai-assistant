#!/bin/bash
awk '
/GemmaModelURL = map\[string\]string{/ {
    print "	GemmaModelURL = \"http://invalid-url-should-not-be-called\""
    skip_until = "}"
    next
}
{
    if (skip_until != "") {
        if ($0 ~ skip_until) {
            skip_until = ""
        }
        next
    }
    print $0
}
' wails-app/backend/classifier/llama_server_test.go > temp.go && mv temp.go wails-app/backend/classifier/llama_server_test.go
