# The numbers on this site were wrong. Here is the refresh.

**Date:** 2026-09-04

I built visa-tracker in 2026 so people like me could look up a salary and a job code and get a straight answer. The seed data I shipped with it used the April 2024 figures: Skilled Worker at £38,700, and `2136` as the software developer code.

Those were not the live rules.

The Home Office had already moved in the 22 July 2025 update. I only caught that when I sat down and checked gov.uk against what the tool was showing. If you used this site before this refresh, it could have told you that you were fine when you were not.

This post is the number change. Why a 2026 site launched on 2024 data is a separate story.

## What I had to change

| Route or code | What the tool had | What it uses now |
|---|---|---|
| Skilled Worker general threshold | £38,700 | £41,700 |
| Scale-up | £36,300 | £39,100 |
| Senior or Specialist Worker (ICT) | £48,200 | £48,500 |
| Immigration Salary List | 80% of going rate, £30,960 floor | 100% of going rate, £33,400 floor |
| Software developers | SOC 2136, £40,000 | SOC 2134, £54,700 |

The Skilled Worker bump is the one most people will feel. Three thousand pounds on the general threshold is already a lot. That last row is worse.

## The bit that actually scared me

In the original seed, `2136` was "programmers and software development professionals" at £40,000. A lot of guides still say that.

That code does not mean that any more.

In the current Appendix Skilled Occupations table, `2136` is IT quality and testing professionals, going rate £41,200. Software developers are `2134`, and the going rate is £54,700 a year on a 37.5-hour week.

So if you typed 2136 into this site and you were a developer, the tool was about £15,000 short. Green when it should have been red.

The rest of the remap is the same story. Cyber security now has its own code, `2135`, at £48,500. DevOps sits under `2139` (IT professionals n.e.c.) at £52,300. Data engineers are under `2133` with architects and business analysts, at £54,900.

If you checked an offer here before this refresh, do it again. Use [Skilled Worker](/visas/skilled-worker) and the code that matches the job you actually have, not the four digits you remember.

## How the salary rule works now

For a standard Skilled Worker application you have to be paid the higher of:

- the general threshold, £41,700 a year
- the going rate for your SOC code

A software developer on 2134 needs £54,700, not £41,700. A tester on 2136 needs £41,700, because the general threshold sits above that code's going rate of £41,200.

The Immigration Salary List used to be the "pay 80% of the going rate" loophole in a lot of people's heads. That is gone. ISL only drops the general floor to £33,400. You still need the full going rate for the job. Almost no core software roles are on the current list. Among the codes I track, biological scientists (2112) and graphic and multimedia designers (2142) are.

## If you have an offer, do this

1. Look up the job on [Job codes](/soc). Search "software", "cyber", or "data". Do not trust a code you wrote down last year.
2. Open the route, usually [Skilled Worker](/visas/skilled-worker), and put in your annual salary plus that SOC code.
3. Green means you meet the higher of the general threshold and the going rate. Amber means you are close, or you might squeeze in via ISL if that code is actually on the list. Red means you are short.

This is a guide, not advice. I am not a lawyer. Before you accept an offer or someone assigns a Certificate of Sponsorship, check the live tables on gov.uk.

## Sources

- [Skilled Worker visa: your job](https://www.gov.uk/skilled-worker-visa/your-job)
- [Appendix Skilled Worker](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-skilled-worker)
- [Appendix Skilled Occupations](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-skilled-occupations) (22 July 2025 going rates)
- [Appendix Immigration Salary List](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-immigration-salary-list)
