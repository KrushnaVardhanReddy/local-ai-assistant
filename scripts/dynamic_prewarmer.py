import asyncio
import json
import sys
import os

# Add backend to sys.path so we can import backend modules
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "backend")))

from llm_client import LLMClient
from local_intelligence import SYSTEM_PROMPT_INJECTIONS, get_local_intelligence
import qa_cache

MOCK_JD = """
Job Title: Senior Frontend Engineer
Company: TechCorp

Responsibilities:
- Build and maintain modern web applications using React, Next.js, and TypeScript.
- Architect scalable, high-performance user interfaces.
- Work closely with designers and product managers to deliver excellent user experiences.
- Optimize web applications for speed and accessibility.
- Mentor junior engineers and participate in code reviews.

Requirements:
- 5+ years of professional experience in frontend development.
- Deep expertise in JavaScript, TypeScript, React, and CSS/HTML.
- Experience with state management libraries (Redux, Zustand, etc.).
- Familiarity with CI/CD pipelines, testing frameworks (Jest, Cypress), and Git.
- Strong understanding of web performance optimization techniques.
"""

MOCK_RESUME = """
John Doe
Frontend Engineer

Experience:
WebSolutions Inc. | Frontend Developer | 2019 - Present
- Led the migration of a legacy Angular app to React, resulting in a 40% reduction in load time.
- Developed reusable UI components used across 5 different products.
- Implemented state management using Redux Toolkit, improving application performance.
- Mentored 3 junior developers and conducted weekly code reviews.
- Improved accessibility score of main product from 75 to 98 on Lighthouse.

StartupX | Junior Web Developer | 2017 - 2019
- Built landing pages using HTML, CSS, and vanilla JavaScript.
- Collaborated with marketing team to A/B test UI designs, increasing conversion rate by 15%.
- Integrated RESTful APIs to display dynamic content.

Skills: JavaScript, TypeScript, React, Next.js, Redux, HTML, CSS, Git, Jest, Cypress.
"""

async def generate_batch(client: LLMClient, batch_num: int, batch_size: int) -> list[dict]:
    prompt = f"""
You are an expert technical interviewer preparing questions and ideal answers for a candidate based on their Resume and the Job Description.

Job Description:
{MOCK_JD}

Candidate Resume:
{MOCK_RESUME}

Formatting rules for answers:
{SYSTEM_PROMPT_INJECTIONS['behavioral']}

{SYSTEM_PROMPT_INJECTIONS['coding']}

Please generate exactly {batch_size} unique interview questions (a mix of behavioral, conceptual, and coding questions) and their ideal answers.
This is batch {batch_num} of 10. Ensure questions are distinct from previous standard interview questions.

You MUST respond in ONLY raw valid JSON array format, with no markdown formatting, no code blocks, and no extra text.
The JSON must be an array of objects, where each object has exactly two string keys: "question" and "answer".

Example format:
[
  {{
    "question": "Tell me about a time you had to optimize the performance of a web application.",
    "answer": "..."
  }},
  {{
    "question": "How does React's Virtual DOM work?",
    "answer": "..."
  }}
]
"""

    messages = [{"role": "user", "content": prompt}]

    # We use stream since it seems the standard way to interact with text LLM models in llm_client
    full_response = ""
    async for chunk in client.stream(messages):
        full_response += chunk

    full_response = full_response.strip()
    # Strip markdown block if present
    if full_response.startswith("```json"):
        full_response = full_response[7:]
    if full_response.endswith("```"):
        full_response = full_response[:-3]
    full_response = full_response.strip()

    try:
        data = json.loads(full_response)
        if isinstance(data, list):
            return data
        else:
            print(f"Error parsing batch {batch_num}: Expected JSON array, got {type(data)}", file=sys.stderr)
            return []
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON in batch {batch_num}: {e}", file=sys.stderr)
        print("Raw response:", full_response, file=sys.stderr)
        return []

async def main():
    print("Initializing LLM client...", file=sys.stderr)
    client = await LLMClient.from_config()

    print("Initializing Local Intelligence engine...", file=sys.stderr)
    get_local_intelligence()

    total_questions = 100
    batch_size = 10
    num_batches = total_questions // batch_size

    all_qa_pairs = []

    for i in range(1, num_batches + 1):
        print(f"Generating batch {i}/{num_batches}...", file=sys.stderr)
        pairs = await generate_batch(client, i, batch_size)
        print(f"Batch {i} yielded {len(pairs)} questions.", file=sys.stderr)
        all_qa_pairs.extend(pairs)

    print(f"Total questions generated: {len(all_qa_pairs)}", file=sys.stderr)

    # Deduplicate before storing to avoid ChromaDB bulk upsert ID collision errors
    unique_pairs = []
    seen_questions = set()
    for pair in all_qa_pairs:
        q = pair.get("question", "").strip()
        if q and q not in seen_questions:
            seen_questions.add(q)
            unique_pairs.append(pair)

    print(f"Total unique questions to store: {len(unique_pairs)}", file=sys.stderr)

    if unique_pairs:
        print("Storing Q&A pairs to ChromaDB cache...", file=sys.stderr)
        stored_count = qa_cache.store_bulk(unique_pairs)
        print(f"Successfully stored {stored_count} pairs to cache.", file=sys.stderr)
    else:
        print("No Q&A pairs generated to store.", file=sys.stderr)

if __name__ == "__main__":
    asyncio.run(main())