# 🔭 Stealth Tech Opportunity Map — Novel Domains

> The core tech: **OS-level invisible window + local Whisper STT + LLM + hotkeys**
> This is a platform, not just one product. Here are 8 domains where it creates a strong wedge.

---

## The Unfair Advantage (Restatement)

Your tech does something **no browser extension or SaaS overlay can do:**
- Window is **truly excluded from OS screen capture** (not just transparent)
- **Runs fully offline** (Whisper STT + local LLM = no data leaves the machine)
- **No cloud dependency** = usable in HIPAA, legal, banking, government environments
- **Hotkeys + voice-activated** = hands-free during live scenarios
- Cross-platform native app (Go/Wails)

Every domain below exploits **at least one** of these as the primary wedge.

---

## Opportunity Scorecard

| # | Product Name | Domain | Competition | Moat | Time to V1 | Revenue Potential |
|---|---|---|---|---|---|---|
| 1 | **StealthPresenter** | Creators / Presenters | 🟡 Medium | 🟢 Strong | 4–6 weeks | $$$ |
| 2 | **ClinicHUD** | Telehealth / Doctors | 🟡 Medium (enterprise heavy) | 🟢 Very Strong | 6–8 weeks | $$$$ |
| 3 | **CounselDesk** | Legal / Attorneys | 🟡 Medium | 🟢 Strong | 6–8 weeks | $$$$ |
| 4 | **MentorGlass** | Coaching / Therapy | 🟢 **Low** | 🟢 Strong | 4–5 weeks | $$$ |
| 5 | **DebateShield** | Students / Debate | 🟢 **Very Low** | 🟡 Moderate | 3–4 weeks | $$ |
| 6 | **AgentPrompt** | Call Center Agents | 🔴 High (Balto, Level AI) | 🟡 Moderate | 8–10 weeks | $$$$$ |
| 7 | **LangShadow** | Language Learners | 🟡 Medium | 🟢 Strong | 5–6 weeks | $$ |
| 8 | **GovBrief** | Govt / PR Spokespeople | 🟢 **Very Low** | 🟢 Very Strong | 4–6 weeks | $$$$ |

---

## 🔍 Deep Dives

---

### 1. 🎤 StealthPresenter — Creators / Presenters
*Already covered in product_analysis.md — this is your primary recommendation.*

**Competition:** ShareSpeak, GhostDesk, FlowPrompter (all dumb scrollers, no local AI)
**Your wedge:** The only one with local AI + voice-tracking + true OS stealth
**Revenue:** $10–$29 one-time or $50/year

---

### 2. 🩺 ClinicHUD — Telehealth / Doctors ⭐ Hidden Gem

**The pain:** Doctors doing telehealth video consults must constantly alt-tab between the patient video, the EHR (Epic/Athena), and their clinical notes. They lose eye contact constantly. Patients notice and feel uncared for.

**The solution:** A stealth overlay pinned near the webcam that shows:
- Patient summary pulled from EHR (via local API hook)
- Relevant drug interactions / dosage lookup
- DSM-5 / ICD-10 code lookup triggered by speech keywords
- SOAP note scaffold auto-generated from the conversation (local Whisper)

**Why competition is weak here:**
- Nuance DAX, Abridge, Suki — all **cloud-based, require BAA contracts**, $300+/month per seat
- None are **visual overlays near the webcam** — they're background audio recorders
- **Zero desktop native tools** exist that do this **completely locally** (HIPAA-friendly by design)

**Why your tech wins:** "No data ever leaves your machine" = instant HIPAA compliance story. This is a massive moat for small/solo practices who can't afford enterprise contracts.

**Revenue:** $49–$99/month per clinician. Even 100 doctors = $60K ARR.

---

### 3. ⚖️ CounselDesk — Legal / Attorneys ⭐ Hidden Gem

**The pain:** Attorneys in virtual depositions, hearings, or client calls need:
- Instant access to case timeline, key exhibits, prior deposition transcripts
- Real-time flagging when witness testimony contradicts prior statements
- Quick statute/case law lookup when opposing counsel references something

**What exists:** Depo Copilot (Filevine), Legal Visor — but they're **cloud-based case management platforms** requiring full CRM buy-in, $400+/month.

**Your wedge:** A **lightweight desktop HUD** that:
- Sits invisible on screen during Zoom/Teams depositions
- Takes a case folder (PDFs, DOCX) on input
- RAG-searches the local corpus in real time when keywords are spoken
- Shows the relevant excerpt next to the camera

**Why it works:** Solo attorneys and small firms (80% of US law market) **cannot afford Filevine**. A $29/month local-first tool that works immediately is an instant yes.

**Revenue:** $29–$79/month. Legal is a high-paying market with low price sensitivity for tools that save billable hours.

---

### 4. 🧠 MentorGlass — Coaching / Therapy ⭐ Lowest Competition

**The pain:** Life coaches, executive coaches, therapists doing video sessions need:
- Client history visible without alt-tabbing
- CBT/REBT framework prompts during tough moments
- Session notes auto-drafted from conversation
- Silence detection (know when to speak vs. hold space)

**Competition:** Nearly ZERO. Therapy note tools (Upheal, Blueprint) exist but are **post-session**. No real-time HUD for coaches exists.

**Your wedge:** A stealth overlay that shows:
- Client profile + previous session summary
- Active technique prompts (e.g., "Use Socratic questioning")
- Auto-drafted SOAP note scaffold building live during session

**Why this is the lowest risk market:**
- No HIPAA complexity for coaches (not covered entities)
- Massive market: 70,000+ life coaches in the US alone, almost none have a real-time AI tool
- Word-of-mouth spreads in coaching communities extremely fast

**Revenue:** $19/month. 500 coaches = $9,500 MRR from a simple V1.

---

### 5. 🏛️ DebateShield — Students / Academic Debate

**The pain:** Competitive debaters (Policy, Lincoln-Douglas, NPDA) need to:
- Quickly reference evidence cards during live rounds
- Track opponent's argument flow and find their prepared responses
- Get AI-suggested rebuttals in real-time

**Competition:** Almost none. Verbatim (an Emacs-based flow tool) is the only real option — and it's ancient.

**Your wedge:** The stealth window isn't needed here (it's in-person), but the **local AI + hotkey + fast RAG search** on their evidence document folder is the core value.

**Who pays:** Debate coaches + speech and debate teams at high schools and colleges. Parents pay for private debate coaching at $100–$200/hour.

**Revenue:** Lower ($9/month student, $49/month coach) but zero competition and passionate niche.

---

### 6. 📞 AgentPrompt — Call Center Agents (Compliance-Grade)

**The pain:** Banking/insurance call center agents must:
- Read mandatory disclosures at specific points in the call
- Avoid saying prohibited phrases ("guaranteed returns", "no risk")
- Instantly look up policy terms when a customer asks a question

**Competition:** Balto, Level AI, Glia — but they're **enterprise SaaS requiring telephony API integration**, $50–$150/seat/month.

**Your wedge:** A local desktop app (no telephony API needed) that:
- Listens via microphone to both sides of the call
- Shows the compliance script at the right moment (keyword-triggered)
- Flags prohibited phrases in real-time

**Why enterprise wins here eventually:** This is a path to $10M+ ARR — but it requires TCPA/FINRA compliance certification, SOC2, and a longer sales cycle. **Not a V1 product** — more like Year 2.

---

### 7. 🌍 LangShadow — Language Learners (Live Calls)

**The pain:** Language learners doing conversational practice on Preply/iTalki or real business calls in a foreign language need:
- Vocabulary lookup without breaking the conversation
- Grammar suggestions shown privately
- Conjugation reminders for tricky tenses
- Translation for words they don't understand in real time

**What exists:** Krisp (audio processing), Google Translate (reactive) — but nothing that sits **invisibly near the webcam** and shows relevant vocabulary as you speak.

**Your wedge:** The stealth HUD shows:
- Words you've been mispronouncing (tracked by Whisper STT)
- Contextual vocabulary suggestions based on what the other person just said
- A "panic mode" hotkey that shows a phrase in your native language

**Revenue:** $9/month — lower, but the TAM is massive (500M+ language learners worldwide). High viral potential in language learning communities (Reddit, YouTube).

---

### 8. 🏛️ GovBrief — Government / PR Spokespeople ⭐ Hidden Gem

**The pain:** Press secretaries, corporate communications officers, and political spokespeople doing live TV/video interviews need:
- Instant fact recall on key statistics during media questions
- Pre-approved talking point access without looking away from camera
- Real-time flagging of "trap questions" based on pre-loaded opposition research

**Competition:** Literally zero software tools. This is done today with paper binders, earpieces ($5K+ broadcast gear), and physical teleprompter hardware.

**Your wedge:** A stealth HUD that:
- Takes a briefing document (PDF) as input
- Listens to reporter questions via Whisper STT
- Surfaces the relevant talking point + key statistic in real time
- Completely invisible in the video feed

**Why it's exciting:**
- No competition in software
- Zero ethical concerns (politicians use teleprompters openly — this is just digital briefing book)
- High willingness to pay: a single enterprise contract (city government, PR firm) = $5K–$20K/year
- The "White House press secretary" demo video writes itself as marketing

---

## 🗺️ Recommended Build Order

```
PHASE 1 (Now) — Product Market Fit
  └── StealthPresenter HUD (creators/salespeople/educators)
      Lowest friction, fastest revenue, best viral hook

PHASE 2 (Month 3–4) — Same Engine, New Audience
  ├── MentorGlass (coaches/therapists)
  │   └── Change UI skin + load client profile format
  └── DebateShield (students/debate)
      └── Remove stealth, keep RAG search + hotkeys

PHASE 3 (Month 6+) — Premium / Enterprise
  ├── ClinicHUD (doctors) — add EHR API hook, HIPAA positioning
  ├── CounselDesk (attorneys) — add case folder RAG, legal corpus
  └── GovBrief (govt/PR) — add briefing doc input, enterprise pricing

PHASE 4 (Year 2) — Big Market Play
  └── AgentPrompt (call centers) — SOC2, telephony integration, enterprise sales
```

---

## The Key Insight

> **You're not building multiple products. You're building one engine with different UX skins and domain-specific prompts.**

The core loop is identical across every product:
1. Load a document corpus (PDF/DOCX/TXT)
2. Listen to speech via Whisper STT
3. RAG-search the corpus based on what's being said
4. Surface relevant results in a stealth overlay near the webcam
5. Gate with Supabase auth + BYOK pass

The only thing that changes per product: **the domain-specific prompt template + the UI skin**.
