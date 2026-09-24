#!/bin/bash
awk '
/^func \(b \*QuestionBuffer\) AddChunk\(chunk string\) \{/ {
    print "func (b *QuestionBuffer) AddChunk(chunk string, ignoreClassifier bool) {"
    next
}
/ctx, cancel := context.WithTimeout\(context.Background\(\), 3\*time.Second\)/ {
    print "		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)"
    next
}
/if !autoFlush {/ {
    print "	if !autoFlush || ignoreClassifier {"
    print "		return"
    print "	}"
    next
}
/return/ {
    if (prev == "if !autoFlush {") {
        # Already handled
    } else {
        print $0
    }
    prev = $0
    next
}
{
    prev = $0
    print $0
}
' wails-app/backend/classifier/buffer.go > temp.go && mv temp.go wails-app/backend/classifier/buffer.go
