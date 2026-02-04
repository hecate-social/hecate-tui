# Why Developer Studio?

*The shortcomings of AI coding assistants, and how we address them.*

---

## The Problem

Current AI coding tools (Cursor, Copilot, Windsurf, Claude Code, etc.) are impressive at generating code snippets but fail at the bigger picture. They're autocomplete on steroids — useful, but not transformative.

They help you write code faster. They don't help you build software better.

---

## Shortcomings & Mitigations

| # | Shortcoming | The Problem | Developer Studio Mitigation |
|---|-------------|-------------|----------------------------|
| 1 | **No domain understanding** | Jumps straight to code without understanding the business domain. Doesn't know what events matter, what the bounded contexts are, what the ubiquitous language is. | **AnD phase** forces domain discovery first — process brief, events, context map, policies. You can't skip to code. |
| 2 | **No architectural guidance** | Happily creates `services/`, `utils/`, `helpers/` and other horizontal patterns. No opinion on structure. | **AnP phase** generates scaffolds following vertical slicing doctrine. Architecture is derived, not debated. |
| 3 | **Context window limits** | Large codebases exceed what fits in the prompt. AI loses track of the big picture. | **Structured artifacts** (DOMAIN.yaml, CONTEXT_MAP.yaml) compress domain knowledge into parseable, focused context. |
| 4 | **No persistent memory** | Each session starts fresh. Forgets decisions, patterns, conventions from yesterday. | **`.hecate/` workspace** persists state, memory, artifacts across sessions. The AI remembers. |
| 5 | **Code-first thinking** | Wants to write code immediately. But code is the last step, not the first. | **Tests first, docs first, inside-out** (InT phase). Code is generated to satisfy tests, not the other way around. |
| 6 | **No project structure awareness** | Puts files wherever. Doesn't understand the intended architecture. Creates drift. | **Screaming architecture** + generated scaffolds enforce structure. Every slice has its place. |
| 7 | **No estimation capability** | Can't tell you how long something will take. No help with planning or budgeting. | **AnP estimation**: sized commands (S/M/L/XL) × count = effort. Planning becomes arithmetic. |
| 8 | **No deployment story** | Helps write code but not ship it. Deployment is "your problem." | **DoO phase**: GitOps for backend, mesh distribution for TUI clients, marketplace for discovery. |
| 9 | **One-shot generation** | Generates code and moves on. No iteration, no refinement, no evolution. | **Phased workflow** with artifacts that persist and can be refined. Iterate at any phase. |
| 10 | **No testing strategy** | Might generate tests if asked, but no systematic approach. Tests as afterthought. | **Inside-out testing doctrine**: state tests → domain tests → integration tests. Systematic, layered. |
| 11 | **Tool agnostic or tool-forcing** | Either no integration with dev tools, or forces you into their IDE/ecosystem. | **Configurable preferences**: use your editor (nvim, code, hx), your tools (lazygit, k9s). TUI launches, you return. |
| 12 | **Hallucination without validation** | Generates plausible but incorrect code. Confident and wrong. | **Tests first** = generated code must pass tests to be accepted. Validation is built in. |
| 13 | **No collaboration model** | Single developer focus. No awareness of teams, reviews, handoffs. | **Artifacts are shareable**: YAML files, diagrams, plans. Version controlled. Review-friendly. |

---

## The Developer Studio Difference

```
Traditional AI Coding          Developer Studio
─────────────────────          ────────────────
"Write me a function"          "Let's discover the domain"
Code first                     Understanding first
One-shot generation            Phased refinement
No structure                   Vertical slicing
No memory                      Persistent workspace
No planning                    Estimates from model
No deployment                  Mesh-native distribution
Autocomplete++                 Software development partner
```

---

## The Four Phases

| Phase | Focus | Output |
|-------|-------|--------|
| **AnD** (Analysis & Discovery) | Understand WHAT | Events, contexts, policies |
| **AnP** (Architecture & Planning) | Plan HOW | Scaffolds, estimates, sequence |
| **InT** (Implementation & Testing) | Build IT | Tests, docs, code (in that order) |
| **DoO** (Deployment & Operations) | Ship IT | Backend + TUI + marketplace |

Each phase produces artifacts. Each artifact feeds the next phase. The AI assists at every step but doesn't skip steps.

---

## Philosophy

1. **Domain first, code last** — Understand before you build.
2. **Structure is doctrine** — Vertical slicing, screaming architecture, no exceptions.
3. **Tests validate, not decorate** — Write tests first, generate code to pass them.
4. **Memory persists** — Sessions end, knowledge remains.
5. **Tools are yours** — We launch your tools, not replace them.
6. **The mesh is the computer** — Distribution is mesh-native.
7. **Planning is arithmetic** — Count the commands, multiply by effort.

---

## Who Is This For?

Developers who:
- Build event-sourced / CQRS systems
- Value domain-driven design
- Want AI assistance without surrendering control
- Prefer terminal workflows
- Need to ship, not just code

---

*The goddess guides the process. You own the product.* 🔥🗝️🔥
