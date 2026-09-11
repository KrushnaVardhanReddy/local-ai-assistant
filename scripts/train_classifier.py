#!/usr/bin/env python3
"""
Train a Logistic Regression classifier on SmolLM2 embeddings.
Run from the project root: python3 scripts/train_classifier.py
"""
import sys, os, pickle
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "backend"))

from local_intelligence import get_local_intelligence
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import cross_val_score
import numpy as np

DATASET = {
    "behavioral": [
        "Tell me about a time you failed.",
        "Describe a situation where you showed leadership.",
        "Give me an example of how you handled a conflict at work.",
        "Walk me through a challenge you overcame.",
        "Tell me about a time you had to learn something new quickly.",
        "Describe a situation where you disagreed with your manager.",
        "Tell me about a time you had to optimize a severely underperforming application.",
        "Describe a situation where you disagreed with a senior engineer about architecture.",
        "Give me an example of a time you led a project under a tight deadline.",
        "Walk me through a time you made a mistake and how you fixed it.",
        "Tell me about a time you had to optimize a severely underperforming React application. What did you do?",
        "Describe a situation where you disagreed with a senior engineer about the architecture of a React component."
    ],
    "coding": [
        "Write a function to reverse a string.",
        "Implement a binary search algorithm.",
        "Write a React hook that debounces a value.",
        "Code a solution for the two-sum problem.",
        "Implement a linked list with insert and delete operations.",
        "Write a function to check if a binary tree is balanced.",
        "Write a custom React hook called useDebounce that takes a value and a delay.",
        "Implement a simple React component that fetches and displays a list of users.",
        "Write a function that flattens a nested array.",
        "Implement rate limiting middleware in Express.",
        "Implement a simple React component that fetches and displays a list of users from an API."
    ],
    "system_design": [
        "How would you design a scalable chat application?",
        "Design the architecture for a URL shortener.",
        "Walk me through designing a distributed caching system.",
        "How would you architect a real-time collaborative editor?",
        "Design a highly available database replication strategy.",
        "How would you design a notification service for millions of users?",
        "How would you design the frontend architecture for a real-time collaborative document editor like Google Docs?",
        "Design a highly scalable e-commerce product listing page with infinite scrolling.",
        "How would you design a global CDN for video streaming?",
        "Walk me through the architecture of a ride-sharing backend.",
        "How would you design the frontend architecture for a real-time collaborative document editor like Google Docs using React?"
    ],
    "conceptual": [
        "What is the Virtual DOM?",
        "Explain how HTTP works.",
        "What is the difference between useMemo and useCallback?",
        "How does garbage collection work in JavaScript?",
        "What is eventual consistency in distributed systems?",
        "Explain the difference between SQL and NoSQL databases.",
        "What is the Virtual DOM in React, and how does it improve performance?",
        "Explain the difference between useMemo and useCallback. When should you avoid them?",
        "What are React Server Components and how do they differ from SSR?",
        "What is the event loop in Node.js?"
    ],
    "opinion": [
        "Do you prefer React or Angular?",
        "What is your opinion on using TypeScript?",
        "Which state management library do you prefer and why?",
        "What do you think about microservices vs monoliths?",
        "Do you prefer SQL or NoSQL for most projects?",
        "What is your take on test-driven development?",
        "Do you prefer using Redux Toolkit, Zustand, or React Context for global state management?",
        "What is your opinion on using Tailwind CSS versus CSS-in-JS libraries?",
        "Do you think Next.js is strictly better than Vite for modern web apps?",
        "What is your preferred approach to API design: REST or GraphQL?",
        "Do you prefer using Redux Toolkit, Zustand, or React Context for global state management, and why?",
        "What is your opinion on using Tailwind CSS versus CSS-in-JS libraries in a large React project?",
        "What is your opinion on using server-side rendering versus client-side rendering?",
        "What is your opinion on using GraphQL over REST APIs?",
        "What is your opinion on functional programming versus object-oriented programming?",
    ],
    "noise": [
        "Okay.",
        "Sounds good, thank you.",
        "Hmm, let me think about that.",
        "Mhm, sure.",
        "Got it.",
        "That's interesting.",
        "Hmm, okay.",
        "That makes sense, thank you.",
        "Sure, I understand.",
        "Let me think for a second."
    ]
}

def main():
    li = get_local_intelligence()
    if not li._enabled:
        print("Error: Local Intelligence (SmolLM2) is not enabled.", file=sys.stderr)
        sys.exit(1)
        
    X, y = [], []
    for label, examples in DATASET.items():
        for text in examples:
            vec = li.encode(text)
            if vec:
                X.append(vec)
                y.append(label)

    if not X:
        print("Error: Failed to encode any examples.", file=sys.stderr)
        sys.exit(1)

    X = np.array(X)
    clf = LogisticRegression(max_iter=5000, C=1.0)
    scores = cross_val_score(clf, X, y, cv=5)
    print(f"Cross-val accuracy: {scores.mean():.2%} ± {scores.std():.2%}")

    clf.fit(X, y)
    out_dir = os.path.join(os.path.dirname(__file__), "..", "backend", "models")
    os.makedirs(out_dir, exist_ok=True)
    out_path = os.path.join(out_dir, "question_classifier.pkl")
    with open(out_path, "wb") as f:
        pickle.dump(clf, f)
    print(f"Saved classifier to {out_path}")

if __name__ == "__main__":
    main()
