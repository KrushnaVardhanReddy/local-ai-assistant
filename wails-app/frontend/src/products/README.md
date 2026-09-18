# Products — Platform Architecture

This directory contains product-specific HUD implementations that share a common
engine and component library.

## How to Add a New Product Skin

1. Create a new folder: src/products/<product-name>/
2. Create <ProductName>HUD.svelte — this is your root component.
3. Import and use ConvPanel and AnswerPanel from ../interview/.
4. Define your headerActions array with product-specific buttons.
5. Write your system prompt in prompts/products/<product-name>/system.txt.
6. Done. No engine changes required.

## Platform-Level vs Domain-Specific Features

| Feature | Level | Where it lives |
|---|---|---|
| Raw/Filter transcript toggle | Platform | ConvPanel (always present) |
| Mic mute toggle | Platform | ConvPanel (always present) |
| Manual mode toggle | Platform | ConvPanel (always present) |
| Copy answer button | Platform | AnswerPanel (always present) |
| Export to file button | Platform | AnswerPanel (always present) |
| STAR Method button | BarnOwl domain | InterviewHUD -> headerActions |
| Catch Me Up | BarnOwl domain | InterviewHUD -> headerActions |
| Hotkeys panel | BarnOwl domain | InterviewHUD -> headerActions |
| SOAP Note output | ClinicHUD domain | ClinicHUD -> headerActions |
| Case Brief | CounselDesk domain | CounselDeskHUD -> headerActions |
| Rebuttal Generator | DebateShield domain | DebateShieldHUD -> headerActions |
