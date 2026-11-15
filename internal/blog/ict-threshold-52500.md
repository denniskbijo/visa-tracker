# The ICT floor is £52,500. This site still said £48,500.

**Date:** 2026-11-15

If you are moving to a UK branch on a Senior or Specialist Worker visa, the route people still call ICT, the general salary floor is £52,500 a year. This site was still showing £48,500.

That is four thousand pounds. An offer in the gap would have looked fine here and failed the rule.

## What changed

[Appendix Global Business Mobility routes](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-global-business-mobility-routes), paragraph SNR 8.1, says the general salary requirement is £52,500 per year, unless SNR 8.3 applies. The pay also has to meet the going rate for the occupation. If the going rate is higher, that is the figure that counts.

[Statement of Changes HC 997](https://www.gov.uk/government/publications/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025-accessible-version) was laid on 1 July 2025. Paragraph APP GBM2 says: in SNR 8.1, for "£48,500", substitute "£52,500". The statement says the changes, other than a short list at the top, take effect on 22 July 2025. An application that used a Certificate of Sponsorship issued before 22 July 2025 is decided under the rules in force on 21 July 2025. The [explanatory memorandum](https://www.gov.uk/government/publications/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025/explanatory-memorandum-to-the-statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025-accessible) says there is no separate transitional arrangement for this salary uprating.

SNR 8.3 is a different, older case. The [eligibility page](https://www.gov.uk/senior-specialist-worker-visa/eligibility) says who that covers. Everyone else should use £52,500, or the going rate if that is higher.

The old £48,500 amount stays in the salary history on the route page. The checker now uses £52,500, effective 22 July 2025.

## How it was caught

The homepage said the data was last checked in September 2026. That line is `last_checked` in the repo, and someone had to edit it by hand. The monthly job opened the gov.uk pages and looked for a few strings we still expect to see. It never wrote a new date. On 15 November 2026 the September date was 64 days old. The limit is 45, so the check failed as stale before it compared the pages.

When the comparison did run, `48,500` was not on the Global Business Mobility appendix. `52,500` was.

## What runs every month now

On the 1st of each month the same check runs again. If every marker is still on the live page, it opens a pull request that sets `last_checked` to that day and does not change a salary figure. If a marker is missing, it opens a GitHub issue listing what is missing, and it leaves the date alone. I still read the appendix and edit the seed.

## If this is your route

Use £52,500 as the general floor, then check the going rate for the job you actually have. You need whichever is higher. Look at [Senior or Specialist Worker](/visas/intra-company-transfer) on this site, then read SNR 8.1 before anyone assigns a Certificate of Sponsorship.

This is a guide, not advice. I am not a lawyer.

## Sources

- [Appendix Global Business Mobility routes](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-global-business-mobility-routes) (SNR 8.1)
- [Statement of Changes HC 997, 1 July 2025](https://www.gov.uk/government/publications/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025-accessible-version) (APP GBM2, in force 22 July 2025)
- [Explanatory memorandum to HC 997](https://www.gov.uk/government/publications/statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025/explanatory-memorandum-to-the-statement-of-changes-to-the-immigration-rules-hc-997-1-july-2025-accessible)
- [Senior or Specialist Worker visa: eligibility](https://www.gov.uk/senior-specialist-worker-visa/eligibility)
