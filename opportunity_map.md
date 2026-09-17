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

> Sorted by **planned build order** — not by revenue. Revenue is only one input; distribution speed, trust barriers, and daily active use all matter equally for a V1.

| 🟢 **Wave 1** | **1A 🎯 PRIMARY** | **BarnOwl AI (Interview Copilot)** | Job Seekers / Tech Candidates | 🔴 Crowded | 🟢 Strong (100% Offline + IDE Workspace) | In Progress | $$$$ | **Active Primary Focus:** Live question interception, real-time code generation, full IDE notes/cheat-sheet workspace, stealth overlay. |
| 🟢 **Wave 1** | **1B ⏸️ Paused** | ~~StealthPresenter~~ | Creators / Presenters | 🔴 Crowded / Low-cost | 🟡 Moderate | Paused | $ | **Paused indefinitely:** ShareSpeak already offers $14.99 one-time lifetime license; teleprompter market is a low-margin commodity race without high-value live AI copilot. MVP preserved in repo. |
| 🟢 **Wave 1** | **2nd** | **MentorGlass** | Coaching / Consulting | 🟢 **Low** | 🟢 Strong | 4–5 weeks | $$$ | Next up: 1 prompt & UI skin swap from BarnOwl IDE, zero competition, B2B monthly SaaS model ($29–$39/mo). |
| 🟡 **Wave 2** | **3rd** | **DebateShield** | Students / Academic Debate | 🟢 **Very Low** | 🟡 Moderate | 3–4 weeks | $$ | No stealth needed, fast build, community-driven distribution |
| 🟡 **Wave 2** | **4th** | **GovBrief** | Govt / PR Spokespeople | 🟢 **Very Low** | 🟢 Very Strong | 4–6 weeks | $$$$ | Zero software competitors, single enterprise deal = huge revenue |
| 🔵 **Wave 3** | **5th** | **CounselDesk** | Legal / Attorneys | 🟡 Medium | 🟢 Strong | 6–8 weeks | $$$$ | High pay, but trust barrier — needs proven track record + testimonials first |
| 🔵 **Wave 3** | **6th** | **ClinicHUD** | Telehealth / Doctors | 🟡 Medium (enterprise heavy) | 🟢 Very Strong | 6–8 weeks | $$$$ | Biggest moat (HIPAA), but needs EHR hooks + compliance story |
| 🔵 **Wave 3** | **7th** | **LangShadow** | Language Learners | 🟡 Medium | 🟢 Strong | 5–6 weeks | $$ | Massive TAM, viral potential, but lower ARPU — best as growth lever |
| 🔴 **Wave 4** | **8th** | **AgentPrompt** | Call Center Agents | 🔴 High (Balto, Level AI) | 🟡 Moderate | 8–10 weeks | $$$$$ | Highest revenue ceiling, but requires SOC2, telephony integration, 6–12mo sales cycle |
| 🟣 **Wave 5 🔭 Explorer** | **Future** | **StreamerHUD** | Twitch/YouTube Streamers | 🟢 Low | 🟢 Strong | 3–4 weeks | $$$ | Invisible overlay for streamers — chat/alerts invisible to OBS. Viral demo potential. 1-prompt swap from StealthPresenter. |
| 🟣 **Wave 5 🔭 Explorer** | **Future** | **LyricsHUD** | Live Singers / Performers | 🟢 Very Low | 🟡 Moderate | 3–4 weeks | $$ | Foot-pedal or timed auto-scroll for lyrics. No STT needed. Same invisible overlay moat. Niche but zero competition. |
| 🟣 **Wave 5 🔭 Explorer** | **Future** | **TabletopDM** | D&D / TTRPG Dungeon Masters | 🟢 Very Low | 🟢 Strong | 4–5 weeks | $$ | LLM panic button for live NPC names, lore lookups, rule clarifications. Massive passionate community. |
| 🟣 **Wave 5 🔭 Explorer** | **Future** | **RTS Co-Pilot** | Competitive RTS Gamers | 🟢 Very Low | 🟡 Moderate | 3–4 weeks | $ | Time-based build order prompter. Very niche, price-sensitive. Low priority but near-zero build effort. |

---

## 🔍 Deep Dives

---

### 1. 🎤 StealthPresenter — Creators / Presenters
*Already covered in product_analysis.md — this is your primary recommendation.*

**Competition:** ShareSpeak, GhostDesk, FlowPrompter (all dumb scrollers, no local AI)
**Your wedge:** The only one with local AI + voice-tracking + true OS stealth
**Revenue:** $10–$29 one-time (V1) → $49/month (V2 with premium features)

#### 📌 V2 Premium Feature: StealthBrowser (Invisible Workspace Panel)

> **Do not build in V1. Revisit in Phase 56.**

Alongside the text HUD, a secondary embedded browser panel that is fully excluded from screen capture (via the same OS-level `SetCaptureExcluded` API). The user sees a live, interactive web app right on their screen; the audience on Zoom sees nothing.

**Legitimate use cases (not cheating):**
- **Sales:** CRM (Salesforce/HubSpot) open invisibly — AI auto-surfaces the account when the client's name is heard.
- **Customer Support:** Zendesk ticket open invisibly — rep never has to say "let me look that up."
- **Financial Advisor:** Portfolio dashboard open invisibly — advisor references live holdings during client calls.
- **Real Estate Agent:** MLS portal open invisibly — AI auto-filters by heard preferences ("3 bedrooms, under $500k").
- **Podcast/Journalist:** Research notes and Wikipedia open invisibly — host never breaks eye contact.

**Why this justifies $49/month vs. $19/month:** It turns StealthPresenter from a teleprompter into a full "invisible workspace" — a category of one.

**Technical note:** Requires launching an embedded Chromium WebView with OS-level window exclusion. Non-trivial but feasible since Wails already uses a WebView internally.

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

**Competition:** Almost none. [Verbatim Paperless Debate](https://paperlessdebate.com) (a legacy VBA macro suite for Microsoft Word, [GitHub](https://github.com/paperlessdebate/verbatim)) is the standard tool used across the circuit for the last 15 years, combined with Excel for manual "flowing". It has zero AI, zero voice tracking, and requires paid desktop Microsoft Word.

**Your wedge:** The stealth window isn't needed here (tournaments are in-person), but the **local offline AI + fast speech-to-text auto-flowing + vector RAG search** across their evidence cards is a 10x leap over Word macros. Because tournaments ban internet, our 100% offline local models are tournament-legal!

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

## 🗺️ Planned Build Order (Revised)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WAVE 1 (Now → Month 3) — Prove the Engine, Get First Revenue
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1A. 🎯 PRIMARY — StealthPresenter HUD
      → Clean ethical positioning, no gray area
      → Individual buyer, viral LinkedIn/Twitter demo
      → Instant revenue signal: validates people pay for stealth HUDs
      → THIS is the flagship product going forward

  1B. ⏸️ PAUSED (Unreleased) — BarnOwl AI (Interview Assist)
      → Engine and code KEPT (reused by hex architecture)
      → Launch is paused indefinitely.
      → Reasons: You rightly noted we haven't even launched this yet! Before going to market, we realized the ethical gray zone + crowded market (Final Round AI, Interview Kickstart, Sensei) makes it a bad *first* launch.
      → Decision: Pivot the marketing and launch to StealthPresenter (1A), using the engine we already built for BarnOwl.

  2nd. MentorGlass (coaches/therapists)
       → 1 prompt swap + skin change on top of StealthPresenter engine
       → Zero competition, word-of-mouth in coaching communities
       → Validates: multi-product skin approach works

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WAVE 2 (Month 3–6) — Expand Reach, Fast Niches
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  3. DebateShield (students/debate teams)
     → Fastest build, no stealth overlay needed
     → Passionate niche, community viral loop

  4. GovBrief (govt/PR spokespeople)
     → Zero software competitors — literal blue ocean
     → One enterprise contract ($5K–$20K/yr) pays for months of runway
     → "White House press secretary" demo = viral press coverage

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WAVE 3 (Month 6–12) — High-Value Professionals
  (Now you have testimonials + proven track record)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  5. CounselDesk (solo/small firm attorneys)
     → $79/month, legal is high-ARPU market
     → WHY NOT EARLIER: Attorneys are conservative buyers.
       They need peer referrals + time to evaluate. Potential malpractice
       liability anxiety around AI-surfaced case law citations.
       You need 3–6 months of proven track record first.

  6. ClinicHUD (telehealth doctors)
     → $49–$99/month, HIPAA-by-design moat
     → Needs EHR API hooks (Epic/Athena) and HIPAA positioning story
     → Requires BAA template + local-only data guarantee

  7. LangShadow (language learners)
     → Massive TAM (500M+ worldwide), viral potential
     → Lower ARPU ($9/month) — better as growth lever than revenue driver

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WAVE 4 (Year 2) — Big Market Play
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  8. AgentPrompt (call center agents)
     → Highest revenue ceiling ($$$$$)
     → Requires: SOC2 Type II, FINRA/TCPA compliance, telephony
       SIP integration, 6–12 month enterprise sales cycles
     → Not first because: Individual agents can't install on locked
       corporate laptops. B2B procurement is slow and expensive.
```

> **Why CounselDesk is Wave 3 and not Wave 1:**
> The revenue is real, but attorneys are the most risk-averse professional
> buyers in existence. A wrong AI citation = potential malpractice. They
> will not adopt an unproven tool from an unknown company.
> By Wave 3, you have StealthPresenter testimonials, MentorGlass reviews,
> and a proven track record — *then* an attorney says yes.

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
