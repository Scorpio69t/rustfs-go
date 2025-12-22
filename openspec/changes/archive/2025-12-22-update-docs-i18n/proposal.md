# Change: Internationalize documentation and comments

## Why
RustFS Go SDK is open source but many comments, docs, and README text are in Chinese, limiting accessibility for global contributors and users. We need English-first docs with a Chinese README variant to reduce confusion and improve adoption.

## What Changes
- Add English-first documentation across repo; translate existing Chinese comments/docs to concise English
- Split README into English default (README.md) and Chinese variant (README.zh.md) with cross-links
- Establish style/coverage rules for doc translations to avoid regressions in future changes

## Impact
- Affects doc files (README) and inline code comments/GoDoc across packages
- Introduces new `docs` specification for internationalization guidelines
- No API or behavior changes; developer-facing documentation only
