# Changelog

## [0.5.0](https://github.com/atbeta/picfast/compare/v0.4.0...v0.5.0) (2026-05-03)


### Features

* **admin:** add two optional footer text and link lines in site settings ([de51829](https://github.com/atbeta/picfast/commit/de51829d05e8dff355336fd368af3854a8bae167))
* **auth:** add first-run setup wizard with gated writes ([9281474](https://github.com/atbeta/picfast/commit/9281474772ff3ed37440daa4d713060681ca5ef5))
* **moderation:** auto-approve pending images when moderation is disabled ([681e96a](https://github.com/atbeta/picfast/commit/681e96a2367d5168f11407331a905080c61729d6))
* **upload:** guest image TTL separate from default ([0e1bce3](https://github.com/atbeta/picfast/commit/0e1bce38ebcdab2db17611de12204fbbd28be81f))


### Bug Fixes

* **admin:** exempt admin uploads from moderation and preserve role on profile update ([632b2e9](https://github.com/atbeta/picfast/commit/632b2e9eb3b5797661ef3f2a850ad3c17769a14c))

## [0.4.0](https://github.com/atbeta/picfast/compare/v0.3.0...v0.4.0) (2026-05-02)


### Features

* **api-tokens:** dual auth and scope enforcement for API tokens ([f1f2dd5](https://github.com/atbeta/picfast/commit/f1f2dd51295d6d9a2c2b761f64c871fc219e2f06))
* **connections:** add flat upload endpoint, redesign connections page with inline token input ([945ed41](https://github.com/atbeta/picfast/commit/945ed416b0136f7246d27ad76d7c8bb2f93f2614))
* **image:** support configurable processing and width variants ([acf4872](https://github.com/atbeta/picfast/commit/acf487240f807de1ac45c7cff9e25dfe32e1e085))
* **mcp:** replace remote MCP server with local @picfast/mcp package ([a138d70](https://github.com/atbeta/picfast/commit/a138d70e3c43cac30dcc56130547f5f184360346))
* **moderation:** add admin review queue and status visibility ([dbab2c9](https://github.com/atbeta/picfast/commit/dbab2c97454d83a7d55f553cc8973ee0e421f12f))
* **settings:** move image processing to user preferences ([c3b5246](https://github.com/atbeta/picfast/commit/c3b5246abed2d3818a2829a65de300f778c3fa78))
* **upload:** add configurable guest quota limits ([9438ceb](https://github.com/atbeta/picfast/commit/9438ceb0d88f5c35dd60729a9a1dd323b5e74cb1))
* **upload:** dynamic image key length with atomic counter ([6875d71](https://github.com/atbeta/picfast/commit/6875d7105abd227a71ae472f59ea7c231dc30f8a))


### Bug Fixes

* **admin-preview:** allow admins to view private thumbnails ([df7278c](https://github.com/atbeta/picfast/commit/df7278c0a2c4876b28798b990e76335383f201a8))
* **observability:** isolate metrics endpoint on internal port ([4c9225d](https://github.com/atbeta/picfast/commit/4c9225dca0afd365d6eb87e8387863debd5dc66d))

## [0.3.0](https://github.com/atbeta/picfast/compare/v0.2.3...v0.3.0) (2026-05-01)


### Features

* **ai-docs:** enrich OpenAPI and MCP semantics for agent usage ([268551d](https://github.com/atbeta/picfast/commit/268551d79082d848695038793edae8daa5a0e9d4))
* **pagination:** add total_pages and page indicator UI ([f243f4b](https://github.com/atbeta/picfast/commit/f243f4b2a9939528e4e4fb787e107a1eacb00029))

## [0.2.3](https://github.com/atbeta/picfast/compare/v0.2.2...v0.2.3) (2026-05-01)


### Bug Fixes

* **docs:** render OpenAPI server URL from runtime base URL ([69be5cd](https://github.com/atbeta/picfast/commit/69be5cd6c9edbb1a73241a9b3a8300e0ed1e3b92))
* **openapi:** allow cross-origin spec imports ([5dc3d68](https://github.com/atbeta/picfast/commit/5dc3d68d8a644b5b388ca69214c31bd7f7252929))

## [0.2.2](https://github.com/atbeta/picfast/compare/v0.2.1...v0.2.2) (2026-05-01)


### Bug Fixes

* **docker:** include OpenAPI spec in runtime image ([9062b54](https://github.com/atbeta/picfast/commit/9062b54f6c9a933133c1bc75ed74b9247ea5fdc9))

## [0.2.1](https://github.com/atbeta/picfast/compare/v0.2.0...v0.2.1) (2026-05-01)


### Bug Fixes

* **i18n:** localize API errors and admin field labels ([0e58a26](https://github.com/atbeta/picfast/commit/0e58a2656222754f7dfa709c4e3d7b1734e54739))

## [0.2.0](https://github.com/atbeta/picfast/compare/v0.1.0...v0.2.0) (2026-05-01)


### Features

* **admin:** expand health and site personalization controls ([36159fc](https://github.com/atbeta/picfast/commit/36159fc6d131d1a216f1bd5a043c490137e9daa1))
* **auth:** enforce verified signup flow ([34e1bcf](https://github.com/atbeta/picfast/commit/34e1bcfa5e9cf5fe1757cdbff78ec2ba99f2ba19))
* **config:** streamline bootstrap and deployment defaults ([992f91e](https://github.com/atbeta/picfast/commit/992f91ebf20ee14417576cab407f16272be9495a))
* **ui:** polish console workflows and album interactions ([2618211](https://github.com/atbeta/picfast/commit/2618211dce5b00b6f3dfa8cbc90294e3ddda23a4))


### Bug Fixes

* **admin:** improve settings safety and SMTP gating ([268091f](https://github.com/atbeta/picfast/commit/268091f7c30de73e0fbb699e8f9d31f887fd8f3d))
* **admin:** track administrator login audits ([012fc8d](https://github.com/atbeta/picfast/commit/012fc8d698ceca7901b5dba998c99782fe6d8ee4))
* **api:** stabilize upload and MCP tool responses ([57ae754](https://github.com/atbeta/picfast/commit/57ae7543495480d2893cf7d2fbe8d0002b532376))
* **i18n:** complete console localization and audit labeling ([db62fc4](https://github.com/atbeta/picfast/commit/db62fc40fbffce208e3f80faa0ceb004edb4c413))
