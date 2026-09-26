package llm

import (
	"fmt"
	"strings"
	"wails-app/backend/session"
)

const MaxContextTurns = 3
const MaxContextTokensApprox = 400

const MockInterviewerPrompt = "You are a senior technical interviewer. " +
	"The user has just provided an answer to a question. " +
	"Evaluate the user's response for accuracy, clarity, and completeness. " +
	"First, give brief, constructive feedback on their answer. " +
	"Then, ask the next relevant follow-up question."

const DefaultSystemPrompt = "You are a stealth interview assistant. The user is in a live technical interview. " +
	"You must provide medium-length, highly structured, and thorough solutions. " +
	"DO NOT repeat the question or the constraints. DO NOT output conversational filler. " +
	"Use bullet points and code snippets where appropriate to fully explain the concept. " +
	"You must format your response strictly as 3 to 5 short, punchy 'Talking Points' bullet points. Do NOT write paragraphs. Make it extremely easy to read at a glance while the user is speaking. Keep each bullet under 15 words. " +
	"EXCEPTION: If the input is a greeting or small talk (e.g. 'hi', 'hello', 'how are you'), respond with a single friendly sentence only — no code. " +
	"If the user asks about their background, resume, or the target job description, you MUST consult the workspace context provided via RAG. The user will upload their resume and job description to the workspace cache."

const DefaultVisionPrompt = "Extract any coding problems, technical questions, or architecture diagrams from this screenshot. Provide a structured approach, pseudocode, and edge cases. Do not write the full code."

const TranscriptSummaryPrompt = `You are a professional meeting note-taker.
The user will provide you with a full transcript of a conversation tagged by speaker.
Your job is to produce:
1. A brief TL;DR (2-3 sentences)
2. Key Discussion Points (bullet list)
3. Action Items (bullet list, if any)
4. Any open questions or follow-ups

Be concise and use professional language.`

var CategoryPromptInjections = map[string]string{
	"behavioral": "This is a behavioral interview question. Use the STAR format strictly:\n" +
		"• Situation (1 short bullet): Set the scene — team size, company stage, or project context.\n" +
		"• Task (1 short bullet): What you were personally responsible for.\n" +
		"• Action (3-5 bullet points): Specific steps YOU took. Use first-person, active verbs.\n" +
		"• Result (1 short bullet): Quantified outcome (%, time saved, revenue impact, etc.).\n" +
		"Do NOT truncate the Action section. Complete all 4 STAR components fully before ending your response.",

	"coding": "This is a coding/algorithmic question. Your response MUST include ALL of the following sections:\n" +
		"1. Approach (1 short bullet): State your chosen algorithm/data structure and why.\n" +
		"2. Code (complete, runnable solution): Write the full implementation — no placeholders like '...' or '# rest of logic here'.\n" +
		"3. Complexity: State Time complexity and Space complexity with a one-line justification for each.\n" +
		"4. Edge Cases (2-3 bullet points): List inputs that could break a naive solution and how your code handles them.\n" +
		"5. Follow-up (optional, 1 short bullet): Mention one optimization or variant the interviewer might ask next.",

	"system_design": "This is a system design question. Structure your answer with ALL of these sections:\n" +
		"1. Clarify Requirements (1-2 short bullets): Scale assumptions, latency SLA, functional/non-functional specs.\n" +
		"2. High-Level Architecture: Key components (load balancer, API gateway, services, DB, cache, queue) with explicit technologies (e.g. PostgreSQL, Redis, Kafka).\n" +
		"3. Data Model: Key entities and relationships.\n" +
		"4. Deep Dive — Critical Path: Walk through the primary request flow end-to-end.\n" +
		"5. Scalability & Reliability: Horizontal scaling, partitioning, replication, failover.\n" +
		"6. Trade-offs: Name 1-2 explicit trade-offs in your design.",

	"conceptual": "This is a conceptual/knowledge question. Your answer must be thorough and structured:\n" +
		"1. Definition (1 short bullet): Precise definition without circular wording.\n" +
		"2. How It Works (3-5 bullet points): Explain the underlying mechanism and WHY, not just WHAT.\n" +
		"3. Code Example (if applicable): Minimal runnable snippet with inline comments.\n" +
		"4. Key Properties / Gotchas (2-3 bullet points): Common misconceptions or edge cases.\n" +
		"5. Real-world Use Case (1 short bullet): Where this concept is used in production systems.",

	"opinion": "This is an opinion/preference question. Give a confident, well-reasoned answer:\n" +
		"1. Position (1 short bullet): State your clear preference or recommendation directly.\n" +
		"2. Reason 1 (1 short bullet): Concrete technical reason grounded in experience.\n" +
		"3. Reason 2 (1 short bullet): Concrete team/process reason.\n" +
		"4. Concession (1 short bullet): Acknowledge when the alternative is better.\n" +
		"Speak in first person. Avoid vague phrases like 'it depends'.",
}

func BuildContextBlock(turns []session.Turn) string {
	if len(turns) == 0 {
		return ""
	}

	var validTurns []session.Turn
	for _, t := range turns {
		if t.InterviewerQuestion != "" || t.AISuggestion != "" {
			validTurns = append(validTurns, t)
		}
	}

	if len(validTurns) == 0 {
		return ""
	}

	// We'll format all turns, and if the total length exceeds 2000,
	// we will start dropping the oldest ones until it fits.
	for {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("=== Recent Interview Context (last %d turns) ===\n", len(validTurns)))
		for i, t := range validTurns {
			qRunes := []rune(t.InterviewerQuestion)
			q := t.InterviewerQuestion
			if len(qRunes) > 120 {
				q = string(qRunes[:120]) + "..."
			}
			aRunes := []rune(t.AISuggestion)
			a := t.AISuggestion
			if len(aRunes) > 200 {
				a = string(aRunes[:200]) + "..."
			}

			sb.WriteString(fmt.Sprintf("[Turn %d] Interviewer: %q\n", i+1, q))
			sb.WriteString(fmt.Sprintf("         AI Answer:   %q\n", a))
		}
		sb.WriteString("=== End Context ===")

		result := sb.String()
		if len(result) <= 2000 || len(validTurns) <= 1 {
			return result
		}
		// trim the oldest turn and try again
		validTurns = validTurns[1:]
	}
}

func BuildSystemPrompt(category string) string {
	if injection, ok := CategoryPromptInjections[strings.ToLower(category)]; ok && injection != "" {
		return DefaultSystemPrompt + "\n\nFor this response: " + injection
	}
	return DefaultSystemPrompt
}

func BuildFullSystemPrompt(category string, turns []session.Turn) string {
	basePrompt := BuildSystemPrompt(category)
	contextBlock := BuildContextBlock(turns)
	if contextBlock == "" {
		return basePrompt
	}
	return basePrompt + "\n\n" + contextBlock
}
