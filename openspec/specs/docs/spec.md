# docs Specification

## Purpose
TBD - created by archiving change update-docs-i18n. Update Purpose after archive.
## Requirements
### Requirement: English-first documentation and comments
RustFS Go SDK SHALL provide English as the default language for README, GoDoc, and inline comments to support global contributors.

#### Scenario: Existing Chinese comments translated to English
- **WHEN** reviewing code comments or GoDoc previously written in Chinese
- **THEN** an English version replaces or accompanies the text
- **AND** the English phrasing is concise and technically accurate
- **AND** any remaining Chinese context is optional and clearly secondary

#### Scenario: New documentation defaults to English
- **WHEN** adding new comments, GoDoc, or docs
- **THEN** the primary content is authored in English
- **AND** maintainers may add Chinese notes only as supplemental context

### Requirement: Bilingual README split with English default
RustFS Go SDK SHALL ship `README.md` in English by default and a dedicated `README.zh.md` for Chinese readers, with mutual links.

#### Scenario: Language switch present in both READMEs
- **WHEN** opening `README.md`
- **THEN** it is written in English
- **AND** it links to `README.zh.md` for Chinese content
- **AND** `README.zh.md` links back to `README.md`

#### Scenario: Content parity across READMEs
- **WHEN** comparing `README.md` and `README.zh.md`
- **THEN** they cover the same sections and instructions
- **AND** the English copy is the source of truth for future updates

