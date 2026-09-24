#!/bin/bash

# 1. Update AskQuestion to zero out lastResponseAt
sed -i '/func (e \*StealthEngine) AskQuestion(question string) error {/a \
	e.mu.Lock()\n	e.lastResponseAt = time.Time{}\n	e.mu.Unlock()' wails-app/core/engine/pipeline.go

# 2. Update handleTranscript to use ignoreClassifier
sed -i 's/e.questionBuffer.AddChunk(cleanTranscript)/ignoreClassifier := time.Now().Before(e.lastResponseAt)\n			e.questionBuffer.AddChunk(cleanTranscript, ignoreClassifier)/g' wails-app/core/engine/pipeline.go

# 3. Add cooldown calculation after LLM finishes
awk '
/finalAns := answerBuilder\.String\(\)/ {
    print $0
    print "			wordCount := len(strings.Fields(finalAns))"
    print "			cooldown := time.Duration(wordCount) * 300 * time.Millisecond"
    print "			if cooldown < 3*time.Second {"
    print "				cooldown = 3 * time.Second"
    print "			} else if cooldown > 20*time.Second {"
    print "				cooldown = 20 * time.Second"
    print "			}"
    print "			e.mu.Lock()"
    print "			e.lastResponseAt = time.Now().Add(cooldown)"
    print "			e.mu.Unlock()"
    next
}
{ print }
' wails-app/core/engine/pipeline.go > temp.go && mv temp.go wails-app/core/engine/pipeline.go
