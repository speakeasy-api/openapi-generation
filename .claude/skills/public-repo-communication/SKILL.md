---
name: public-repo-communication
description: Use before writing anything that lands in this repository or on GitHub — PR titles/bodies, comments and review replies, commit messages and trailers, changesets, branch and file names, code comments and string literals, fixtures/specs/tests, docs, screenshots. Public text must describe the change on its own terms and never carry customer-derived identifiers, private paths, internal references, or workflow provenance.
---

# Public-repo communication

This repository is public and generates SDKs, CLIs, MCP servers, and Terraform
providers for many organisations. Anything written here — or on a GitHub PR,
issue, review, or comment — is readable by everyone, indexed, mirrored, and
outlives the conversation that produced it. It therefore has to **stand
alone**: a maintainer with zero context should understand what changed and why,
and no reader should learn who our customers are or how our internal work is
organised.

## The rule

1. **Never publish customer or customer-derived identifiers.** Not the company,
   product, API, spec, or generated CLI/SDK/package name; not their operation
   ids, command names, schema/property/enum names, hostnames, header names,
   model ids, or verbatim error strings/payloads that only that API produces;
   not "the customer" (or "a user", "a partner") plus details that narrow it
   to one; not links to their material even when it is public. Public
   discoverability does not make a customer name safe here. Never publish
   credentials, customer documents, or unredacted private logs, even when
   they carry no customer identifier.
2. **Never expose internal context.** No private repository URLs or names, no
   real local filesystem paths (`/Users/...`, `~/...`) or workspace/worktree
   names copied from an environment (repository-relative paths are fine), no
   private-tracker ids or links, no session/thread ids, no internal document
   or spec titles, no internal reviewers or tools named as process actors.
3. **No workflow provenance.** Do not describe how the work was organised —
   which agent, model, reviewer round, session, sibling branch, or orchestrator
   produced a finding or decision. State the finding as a fact about the code.
4. **Describe the change, not the story.** Say what the code does now, what it
   did before, and why the new behaviour is correct. If motivation matters,
   describe the *API shape* that motivates it (see below), never its owner.

Applies to humans and AI agents equally, and to every public artifact and its
metadata: PR title/body, PR/issue/review comments and replies, labels, commit
messages and trailers (`Co-authored-by` included — commits become public on
push and survive rebases and squashes), branch and tag names, `.changesets/`,
CHANGELOG, code comments, string literals and templated error text, file/
fixture/spec/schema/test names, sample values, generated review SDK output,
docs, screenshots and attachments.

## Rewrites

| Don't | Do |
| --- | --- |
| "The FooCorp CLI needs `--dry-run` on every command" | "CLIs generated for APIs with keyless preview endpoints need `--dry-run` on every command" |
| "FooCorp's spec returns 400 with `API_KEY_INVALID`" | "An API that reports invalid credentials as HTTP 400 with a reason code in the body" |
| "Tested against the foocorp-cli-poc reference" / "the FooCorp fixture" | "Tested against a hand-written CLI with the target UX" / "the review fixture" |
| "Fixed in `/Users/me/Code/my-worktree`" | omit entirely |
| "codex round B found the retry loop double-counts" | "The retry loop double-counted attempts when `Retry-After` was present" |
| "the exit-codes agent's PR" / "sibling PR from the async worker" | "PR #70" / "the PR that adds `--async`" |
| "per orchestrator decision / per <teammate>'s session" | "Decision: …" or just state the behaviour |
| "see TICKET-1234 / internal doc '<Spec title> rev 1.0' item A8" | describe the requirement inline, or link a public issue |
| fixture `foocorp-models.yaml` with `api.foocorp.com`, `foocorp-pro-2` | `route-dispatch.yaml` with `api.example.com`, `model-a`; schemas `Widget`, `Pet`, `AcmeOrder` |
| PR example copied from a customer overlay (`<their-command>`, `<TheirOp>#<TheirVariant>`) | the same example expressed with the repository fixture's names |

Allowed: the repository's own project, targets, features and public
repositories; public open-source dependencies, standards and tools (`cobra`,
`zod`, `httpbin`, RFC 9110, OpenAPI 3.1, `google.rpc.ErrorInfo` as a
convention, GitHub Actions); public GitHub PR/issue numbers and SHAs;
@-mentioning reviewers by their public GitHub handle. A customer's product,
repository, docs link, or SDK is *not* covered by this carve-out even when it
is public.

## When context genuinely matters

Sometimes a change only makes sense because of a real API's behaviour. State
the minimum abstract shape needed to review and test it, normalise every
non-essential identifier and literal, and broaden or split rare combinations
that would fingerprint one API:

- "an API whose list endpoints return a `nextPageToken` but no total count"
- "an operation whose 200 response is a stream when `Accept: text/event-stream`"
- "an auth scheme that accepts either an API key header or OAuth2 bearer"

That is enough for a reviewer to judge the change and for tests to cover it.
The owner of the API is never load-bearing. If the behaviour cannot be
explained without a source-specific name, value, or link, do not publish it —
ask privately for a safe abstraction.

## Pre-publish checklist

Run this over the exact artifact you are about to post or commit (body,
comments, commit messages, branch/file names, the full staged diff); inspect
screenshots and attachments by eye — text search does not cover them:

- [ ] Enumerate the customer/product/spec names, model ids, hostnames, command
      and operation names from the private inputs used for this task (from the
      source material — do not rely on memory); `grep -Fi` each literally: zero
      hits, except values independently present in repository fixtures or
      public standards.
- [ ] `grep -E '/Users/|~/|/home/'` plus your actual workspace/worktree names:
      zero real-path hits.
- [ ] `grep -Ei 'agent|codex|orchestrator|worktree|round [a-z0-9]|session|sibling|thread'`
      — review each remaining hit semantically; keep only technical meanings
      (the documented "agent mode" feature, an HTTP session, a sibling schema
      node, a public review thread) and remove anything that attributes work or
      reveals private coordination.
- [ ] No private tracker ids, private links, or real request/trace/session
      ids. Clearly synthetic UUIDs in fixtures and public PR/issue numbers are fine.
- [ ] Read the summary as a stranger: is it clear what changed and why without
      any conversation context? Are findings and decisions stated directly —
      no attribution or process narration, even where no keyword matched?
- [ ] Fixtures, specs, sample values, and test names use neutral names
      (`acme`, `petstore`, `widgets`; reserved domains `example.com`, `.test`,
      `.invalid`).
- [ ] The PR template's public-safety checklist is honestly ticked.

If a check fails, rewrite before publishing. If text was already published,
edit it in place (PR body, comment) or reword the commit before it merges —
and remember edits do not erase notifications, forks, or caches; rotate any
credential that leaked.

Related: `CONTRIBUTING.md` (repository-wide rules this skill turns into a
writing procedure) and `.github/pull_request_template.md` (per-PR checklist).
