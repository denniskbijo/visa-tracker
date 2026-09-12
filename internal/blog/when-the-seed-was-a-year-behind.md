# I launched a 2026 site on 2024 visa numbers. Here is what I learned.

**Date:** 2026-09-06
**Status:** draft, not on `/blog` yet

I built visa-tracker because I was tired of jumping between gov.uk pages, salary tweets, and half-updated blogs every time someone asked "can I get a Skilled Worker on this offer?" I wanted one place for people like me: tech and STEM, any nationality, no account, no sales pitch.

So I put the usual stuff in one site. You can compare the main work routes, search the licensed sponsor list, look up a SOC code, run a short wizard, and check a salary against the threshold. The sponsor register already pulls from gov.uk. The salary floors and job codes I stored in the repo myself.

I launched it in February 2026. In September I found out those stored numbers were a year behind. I wrote up [what I changed](salary-thresholds-2025.md). This is how they got there.

## February 2026

When I filled the salary and SOC seed, I did it with a coding assistant. It did not open Appendix Skilled Occupations. It wrote down what it already "knew":

- Skilled Worker general threshold: £38,700 (the April 2024 rise everyone still talks about)
- Scale-up: £36,300
- ICT: £48,200
- Software developers: SOC `2136` at £40,000

Those numbers were right in 2024. By the time the site went live, the Home Office had already published the 22 July 2025 update. The general floor was £41,700, and programmers had moved to `2134` at £54,700.

I did not catch it. The figures looked familiar, the kind of thing you have seen in a dozen articles, so I let them through.

## September 2026

Seven months later I asked a simple question: is this still the UK tech visa landscape, or are we missing something?

This time the assistant went to gov.uk. I sat with the appendices and read them myself. The going rates, the job codes, the ISL floor, they all matched what it had just pulled. I could point at a row on the official page and see the same number in the tool.

That is what changed for me. The first pass used memory. This pass used the source.

## What I am taking from this

If the official page exists, open it. A model, a blog, even your own memory can sound sure about a number that decides whether someone is eligible.

In this codebase I want a simple habit. The sponsor list already comes from the live CSV, so leave that alone. Thresholds and SOC codes can stay in YAML, as long as the file says when it was taken and which appendix it came from. If that date is old, treat the file as a snapshot until someone re-reads gov.uk. And anything an assistant writes into `data/` gets checked against the live page before it ships.

I am not naming tools here. I care about the difference between "this sounds right" and "this is on the page today."

I built this so I would have somewhere I trusted with my own offer. That only works if I keep going back to the source. If this refresh stops one person taking a green light that should have been red, that is enough reason to do it again the next time the rules move.
