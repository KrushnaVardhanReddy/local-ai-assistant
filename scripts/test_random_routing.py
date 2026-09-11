import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "backend"))
from local_intelligence import get_local_intelligence

TESTS = [
    # Behavioral
    "Could you share an experience where you had to quickly adapt to a major change in project scope?",
    "I'd love to hear about a time you mentored a junior developer.",
    "Has there ever been a time you missed a critical deadline? What happened?",
    "Walk us through a scenario where you successfully resolved a conflict within your team.",
    
    # Coding
    "Can you quickly whip up a Python decorator that logs function execution time?",
    "We need to sort an array of objects by a specific property. How would you code that in JS?",
    "Could you sketch out an algorithm to find the longest palindromic substring?",
    "Write a SQL query to find the second highest salary from an Employee table.",
    "Build a custom hook to manage a WebSocket connection.",

    # System Design
    "Suppose we want to build a clone of Twitter. What would the high level architecture look like?",
    "How would you approach scaling our Postgres database to handle 10x current traffic?",
    "Design a rate limiter for a public API.",
    "If you were architecting a live streaming service, what protocols and components would you use?",
    
    # Conceptual
    "Could you explain the difference between a process and a thread?",
    "What exactly is event bubbling in the DOM?",
    "Can you describe how consistent hashing works?",
    "What's the difference between an abstract class and an interface?",
    "How does the V8 garbage collector decide what to free?",
    
    # Opinion
    "Are you a fan of using ORMs, or do you prefer writing raw SQL?",
    "What are your thoughts on monolithic vs micro-frontend architectures?",
    "Do you prefer strongly typed languages or dynamically typed ones?",
    "Which CSS framework do you think is best for a small startup?",
    
    # Noise
    "Yeah, that makes total sense.",
    "Oh, I see what you mean.",
    "Wait, could you repeat the first part?",
    "Mhm.",
    "Interesting...",
]

def main():
    li = get_local_intelligence()
    print("\n=== RUNNING 27 RANDOM TESTS ===\n")
    for q in TESTS:
        cat = li.classify_question(q)
        print(f"[{cat.upper():<13}] {q}")

if __name__ == "__main__":
    main()
