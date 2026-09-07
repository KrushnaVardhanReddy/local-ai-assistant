# Adversarial Testing

Project Parakeet (Local AI Assistant) is developed in an ongoing arms race with an internal sister project: the **Interview Integrity Platform**. 

## The Dynamic
While Parakeet is designed to be the ultimate stealth interview copilot, the Interview Integrity Platform is designed to catch tools exactly like it. 
- **Parakeet's Stealth**: Uses OS-level flags (e.g., `WDA_EXCLUDEFROMCAPTURE` on Windows) to make its UI invisible to screen shares.
- **The Integrity Platform**: Uses a Layer 2 Native Rust Agent to enumerate OS windows and detect the presence of capture-exclusion flags, even if it cannot see the window contents.

## Goal
The goal of testing Parakeet against the Integrity Platform is to ensure Parakeet remains completely undetectable. We want the Integrity Platform to **"catch others, not ours."** 

By controlling both the sword and the shield, we can find the absolute bleeding edge of stealth technology. If the Integrity Platform can detect Parakeet's invisible window, we will evolve Parakeet (e.g., relying purely on Remote Helper Mode on a secondary device) to bypass it entirely.
