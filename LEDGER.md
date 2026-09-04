# Ledger

Working list for visa-tracker. Pick one item, ship it, mark it done. Do not start a second item until the first is committed.

Source for figures: gov.uk Immigration Rules (Appendix Skilled Worker, Appendix Skilled Occupations, Appendix Immigration Salary List). Seed snapshot in `data/` is the 22 July 2025 update unless a later date is noted.

## Done

- [x] Eligibility checker on `/visas/{slug}` (green / amber / red vs threshold)
- [x] Visa wizard (`/wizard`)
- [x] Personal timeline (`/timeline`), including JSON embed fix
- [x] Data refresh (Sep 2026): SW £41,700, Scale-up £39,100, ICT £48,500; SOC remap (2134 programmers at £54,700); ISL is 100% going rate + £33,400 floor

## Next

- [ ] Blog post on the data refresh: what changed in 2025/26, why old SOC 2136 was wrong for developers, and how to re-check an offer
- [ ] Eligibility extras: new-entrant option (£33,400 + 70% going rate), STEM/non-STEM PhD discounts, £17.13 hourly floor
- [ ] Route comparison table: pick 2–3 routes, compare sponsor, salary, duration, ILR, English, extendable
- [ ] Expand SOC set further (more STEM Table 1 codes; optional Temporary Shortage List flag)
- [ ] Salary calculator: job title → SOC match → minimum salary vs offer

## Data and automation

- [ ] Processing times: refresh beyond the March 2025 YAML snapshot (scraper or manual)
- [ ] Policy-change banner when seed thresholds or rules change after ingest
- [ ] Richer sponsor results: explain ratings, link to gov.uk, suggest next steps
- [ ] Stop appending duplicate processing-time rows on every startup (thresholds already skip duplicates)

## Product copy and routes

- [ ] Wizard: English B2 for first-time Skilled Worker, Scale-up, HPI
- [ ] HPI: university list size, annual cap, PhD = 3 years (description updated; wizard/timeline still treat all HPI as 2 years)
- [ ] Graduate: 18-month grant from 1 Jan 2027 for most non-PhD applicants
- [ ] Innovator Founder route (tech founders)
- [ ] Global Talent: Talent vs Promise ILR (3 vs 5 years) in timeline

## Growth and polish

- [ ] Smarter timeline: multi-visa history, route switches, more accurate ILR rules
- [ ] Country-specific guides (community)
- [ ] Onboarding tour: job code → salary → sponsors
- [ ] Mobile polish for tables and comparison views
- [ ] Handler, store, and ingest tests beyond wizard/eligibility
- [ ] Content-Security-Policy header

## Infrastructure

- [ ] Pin gosec to a release tag (CI currently uses `@master`)
- [ ] Upgrade CodeQL action to v4
