.. SPDX-License-Identifier: AGPL-3.0-or-later

GoreeCloud AI Contribution Policy
=================================

Purpose
-------

GoreeCloud Search is developed in an environment where artificial-intelligence
agents and AI-assisted engineering are intentional parts of the software
lifecycle. This policy defines how AI-assisted and agent-authored contributions
are accepted in this GoreeCloud-maintained fork without weakening correctness,
security, privacy, review, provenance, licensing, recovery, or production
controls.

Governing principle: AI participation is permitted. Quality, evidence,
accountability, and authorization determine whether a contribution is
acceptable; AI authorship by itself does not make a contribution invalid.

Permitted AI and Agent Use
--------------------------

- **AI-assisted and agent-authored contributions are permitted.** Contributors
  may use coding assistants, language models, autonomous development agents,
  local models, hosted models, or other AI-supported engineering tools.
- **AI agents may be the primary implementation author.** GoreeCloud does not
  require a human-written majority of the code, documentation, tests, pull
  request description, or review discussion merely because AI was involved.
- **Agent work must use the normal repository workflow.** Material changes
  should remain attributable to branches, commits, pull requests, automated
  validation, review records, and other applicable source-control evidence.
- **The same engineering requirements apply regardless of authorship.** Human,
  AI-assisted, and agent-authored changes remain subject to the same applicable
  product, security, privacy, testing, documentation, upstream-maintenance,
  licensing, release, and recovery controls.

Transparency and Accountability
-------------------------------

Material AI or agent involvement should be identified in the pull request or
other repository record when that information is useful for auditability. A
concise statement naming the tool, agent, or type of assistance is sufficient
when the exact system is known.

A disclosure is not a substitute for technical evidence. Contributors and
agents remain responsible for ensuring that claims in a pull request are
accurate and supported. Do not fabricate test results, reviews, deployment
results, release state, security findings, runtime evidence, or other completed
work.

No contribution may be automatically rejected or closed solely because AI was
the primary author, because an AI system helped write the pull request
description or discussion, or because a particular AI-policy checkbox was not
present. Review should evaluate the contribution itself and the evidence that
supports it.

Quality and Review
------------------

AI use does not lower the review bar. Contributions should be coherent,
maintainable, scoped to the intended change, and supported by the applicable
validation. Low-quality, misleading, unreviewable, unsafe, or unvalidated work
may be rejected for those substantive reasons regardless of whether it was
written by a person or an AI system.

When a change materially affects security, privacy, shared architecture,
integrations, release state, or production behavior, the applicable specialized
review and acceptance gates remain required. An implementation agent must not
silently treat its own successful code generation or local test result as
production approval.

Security, Privacy, and Credentials
----------------------------------

AI-assisted development must preserve GoreeCloud security and privacy
boundaries.

- Do not place secrets, private keys, reusable credentials, protected user data,
  private infrastructure information, or other sensitive GoreeCloud material in
  source, public discussions, ordinary documentation, CI logs, or unapproved AI
  prompts and services.
- Repository, workflow, environment, package, deployment, and agent permissions
  must follow least privilege.
- Untrusted pull-request code must not automatically receive privileged
  production credentials.
- Security-sensitive automation must fail closed when repository identity,
  permissions, exact revision, or required evidence is materially ambiguous.

Lifecycle and Production Authority
----------------------------------

AI agents may implement changes, write tests, run validation, create pull
requests, respond to review, update documentation, build artifacts, and prepare
release candidates when authorized. Those activities do not automatically grant
production authority.

Source implementation, CI success, pull-request merge, release-candidate
creation, release publication, deployment, production acceptance, and Stable
lifecycle approval are separate states. High-risk or production-impacting
decisions remain subject to the applicable GoreeCloud approval and acceptance
requirements.

Upstream SearXNG Boundary
-------------------------

GoreeCloud Search is a maintained fork of SearXNG. This policy governs
contributions to the GoreeCloud repository. It does not override the contribution
or AI policies of the upstream SearXNG project.

When a change is submitted directly to upstream SearXNG, contributors must
follow the current upstream project's rules for that upstream submission. A
GoreeCloud agent may prepare an upstream-capable change, but the upstream project
retains authority over whether and how it accepts that contribution.

Final Policy
------------

GoreeCloud welcomes disciplined AI-assisted and agent-authored engineering.
Automation should reduce repetitive work while keeping the repository more
understandable, testable, auditable, secure, private, recoverable, and under
GoreeCloud control.
