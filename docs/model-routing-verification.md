# Chat model routing verification

Verified on 2026-11-12 against the live Gemini web upstream using an anonymous
session and this repository's request payload format. The prompt was
`Reply with exactly: ROUTE_OK`.

Previously, the registry advertised several model versions and response handlers
echoed the requested name. Several names shared the same request payload, and
unknown names silently fell back to the default. A successful response therefore
did not establish which model served it.

## Observed upstream metadata

The model label came from index 42 of the inner `wrb.fr` response payload; the
response selector came from index 58. Generated self-identification text was not
used as evidence.

| Requested selector | Former registry route | HTTP status | Upstream label | Response selector |
| --- | --- | --- | --- | --- |
| 0 | Argon | 200 | `3.5 Flash-Lite` | 6 |
| 1 | Flash 3.5 / 3.6 / 3.7 / 3.8 | 200 | `3.5 Flash-Lite` | 6 |
| 2 | Thinking | 200 | `3.5 Flash-Lite` | 6 |
| 3 | Pro | 200 | `3.5 Flash-Lite` | 6 |
| 4 | Auto | 200 | `3.5 Flash-Lite` | 6 |
| 5 | Thinking Lite | 200 | `3.5 Flash-Lite` | 6 |
| 6 | Flash Lite | 200 | `3.5 Flash-Lite` | 6 |
| 3, extra fields 31=2 and 80=3 | Pro-enhanced | 200 | `3.5 Flash-Lite` | 6 |

Selector 1 succeeded after a transport retry. These observations apply to the
tested anonymous session. Signed-in accounts, entitlements and future upstream
changes can produce different routing. This is a normalized Gemini web label,
not a claim that the name is an official Google API model ID.

## Resulting behavior

- Chat model lists contain only `gemini-3.5-flash-lite`, routed with selector 6.
- Removed names and unknown chat models return HTTP 400.
- Chat completions, Responses and Gemini-format output derive their model name
  from upstream metadata. Missing metadata yields `gemini-web`; a role-only
  opening stream chunk uses that name until metadata arrives.
- Existing deployment configuration that sets an old `DEFAULT_MODEL` or
  `default_model` must be updated to `gemini-3.5-flash-lite`.
- The separate image endpoint and its configured backend remain available.

## Validation

The Go build and the full local test suite passed. Local checks covered the
single-model lists, rejection of removed names, metadata parsing, and model
reporting across the three response formats. Live gateway calls also passed for
normal chat completions and SSE streaming, reporting `gemini-3.5-flash-lite` and
the expected concatenated text `ROUTE_OK`.

Test additions and fixture changes remain local and are excluded from the
commit. No deployed gateway URL was supplied, so production deployment was not
verified.
