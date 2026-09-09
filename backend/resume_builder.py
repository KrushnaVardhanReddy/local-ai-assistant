from llm_client import LLMClient

async def generate_tailored_resume(llm_client: LLMClient, base_resume: str, job_description: str) -> str:
    """
    Takes a user's base resume and the target Job Description, and uses the LLM
    to generate a perfectly tailored resume formatted in raw Markdown.
    It optimizes bullet points to match the keywords in the Job Description
    without hallucinating fake experience.
    """

    if not llm_client:
        raise RuntimeError("LLM Client not initialized")

    system_prompt = (
        "You are an expert resume writer and career coach. Your task is to tailor the candidate's base resume "
        "to perfectly match the provided Job Description. "
        "Output ONLY raw Markdown. Do not wrap the output in markdown code blocks. "
        "Do not include any conversational text or preambles. "
        "Optimize the bullet points to naturally incorporate keywords from the Job Description. "
        "DO NOT hallucinate or invent fake experience, skills, or education. Only emphasize and rephrase "
        "what the candidate actually has based on the base resume. Keep the format professional and clean."
    )

    user_prompt = f"""--- BASE RESUME ---
{base_resume}

--- TARGET JOB DESCRIPTION ---
{job_description}
"""

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_prompt}
    ]

    result = ""
    try:
        async for token in llm_client.stream(messages):
            result += token
    except Exception as e:
        raise RuntimeError(f"LLM Error: {str(e)}")

    # Strip markdown code fences if LLM wrapped the response anyway
    result = result.strip()
    if result.startswith("```markdown"):
        result = result[11:]
    elif result.startswith("```"):
        result = result[3:]

    if result.endswith("```"):
        result = result[:-3]

    return result.strip()
