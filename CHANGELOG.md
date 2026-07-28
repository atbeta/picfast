# Changelog

## [0.22.0](https://github.com/atbeta/picfast/compare/v0.21.0...v0.22.0) (2026-07-28)


### Features

* **cli:** add config commands and harden Windows installs ([141336b](https://github.com/atbeta/picfast/commit/141336b6a54b6807b0100dd6ce8c49689075b40c))


### Bug Fixes

* **audit:** record image.upload for sharex and flat upload endpoints ([d1bf3f1](https://github.com/atbeta/picfast/commit/d1bf3f172227ce49208f6b1ce48cfca192d27edf))
* **cli:** make npx picfast resolve the upload binary ([f82082f](https://github.com/atbeta/picfast/commit/f82082f1017cd64c49cb98cfdde7f9e303af42c9))
* **web:** jump back a page after deleting the last item(s) on a page ([a97bf45](https://github.com/atbeta/picfast/commit/a97bf45bda6f157f3790930b03a608256e10dce0))

## [0.21.0](https://github.com/atbeta/picfast/compare/v0.20.2...v0.21.0) (2026-07-24)


### Features

* **web:** add Ctrl+V paste-to-upload on upload pages ([69ca3b4](https://github.com/atbeta/picfast/commit/69ca3b462ab3f231e9dc20ded8c6721e95f00e3a))

## [0.20.2](https://github.com/atbeta/picfast/compare/v0.20.1...v0.20.2) (2026-07-24)


### Bug Fixes

* UI consistency and i18n/audit gaps ([9ac69e3](https://github.com/atbeta/picfast/commit/9ac69e3a4799aed85d6b9512acfec76115e47bd4))

## [0.20.1](https://github.com/atbeta/picfast/compare/v0.20.0...v0.20.1) (2026-07-24)


### Bug Fixes

* **spa:** use no-store for index.html to prevent stale cache after deploy ([7816875](https://github.com/atbeta/picfast/commit/7816875bfaf73135a32b828bae6a512e6d2ce14f))
* **web:** add i18n labels for new audit actions and resource types ([c7bc826](https://github.com/atbeta/picfast/commit/c7bc8266b82514ab170932f2fabd4377c0c5fd71))
* **web:** match new-tag button size to other create buttons ([5c7b17d](https://github.com/atbeta/picfast/commit/5c7b17db75a5e3462725ce6a17cb4627cd27d901))

## [0.20.0](https://github.com/atbeta/picfast/compare/v0.19.0...v0.20.0) (2026-07-24)


### Features

* **asset:** tags, EXIF/OCR display, advanced search, and i18n ([c970852](https://github.com/atbeta/picfast/commit/c970852e3b469266de8f2168bac4871511f00c40))
* **audit:** enable upload logging by default; add 7 new audit events ([450ccf6](https://github.com/atbeta/picfast/commit/450ccf670b9b56e1e84c42dbb944add5566a735e))
* **cli:** add picfast CLI upload client for npm distribution ([342a56c](https://github.com/atbeta/picfast/commit/342a56c48461885225bc704b364b47ba84b3b976))
* **site:** 4 admin toggles for white-label internal deployments ([3d99120](https://github.com/atbeta/picfast/commit/3d99120627229c9f63cd4f0d7693281ec6d3b99b))
* **vscode:** add PicFast Image Uploader VS Code extension ([de5621d](https://github.com/atbeta/picfast/commit/de5621dd3927bdd9239664c3074dc7f73080f72d))
* **web:** replace third-party VS Code exts with official PicFast extension ([bd7c518](https://github.com/atbeta/picfast/commit/bd7c518c6f0d1146e1551f1836d539d07ef14e78))


### Bug Fixes

* **ci:** restrict docker-publish tag filter to version tags ([b320de2](https://github.com/atbeta/picfast/commit/b320de20587b8d3dc52546914040b2cb3bbaea1c))
* **vscode:** remove $(cloud-upload) from notification; localize CodeLens title ([298e8b2](https://github.com/atbeta/picfast/commit/298e8b255d7d64d6262d0cdc0b10d9e1dd51e78b))
* **vscode:** revert default baseUrl to safe placeholder ([b3cf901](https://github.com/atbeta/picfast/commit/b3cf9012e80ded0d79a3fb99075c3722c69b8055))

## [0.19.0](https://github.com/atbeta/picfast/compare/v0.18.2...v0.19.0) (2026-07-24)


### Features

* **app:** configurable image key min length ([b5beef3](https://github.com/atbeta/picfast/commit/b5beef389d7159ea35223884af4635150fc07432))

## [0.18.2](https://github.com/atbeta/picfast/compare/v0.18.1...v0.18.2) (2026-07-11)


### Bug Fixes

* **spa:** cache hashed assets 1y immutable, revalidate index.html ([f0cf4c4](https://github.com/atbeta/picfast/commit/f0cf4c4171233db3b7a314f7f52cc421235f6222))

## [0.18.1](https://github.com/atbeta/picfast/compare/v0.18.0...v0.18.1) (2026-07-11)


### Bug Fixes

* **spa:** cache hashed assets 1y immutable, revalidate index.html ([f0cf4c4](https://github.com/atbeta/picfast/commit/f0cf4c4171233db3b7a314f7f52cc421235f6222))
* **web:** pin vite to 8.0.x to avoid rolldown CJS wrapper regression ([7d82de5](https://github.com/atbeta/picfast/commit/7d82de549fcc4df9a87add00a1aac4a9f00fa0cf))

## [0.18.0](https://github.com/atbeta/picfast/compare/v0.17.1...v0.18.0) (2026-07-11)


### Features

* **ci:** auto-translate release notes to Chinese via LLM ([37adb6d](https://github.com/atbeta/picfast/commit/37adb6d95024618ed62c4154d756eca3f128346e))
* **images:** add EXIF extraction and perceptual hash (pHash) ([a3d5bf6](https://github.com/atbeta/picfast/commit/a3d5bf6c8cf7fe284b9c4ee0ada90e6a7d6242d6))
* maintenance page overhaul, cross-platform disk, and review fixes ([fc78bdd](https://github.com/atbeta/picfast/commit/fc78bddb948d287f9bfb39e31169f9cd0fbbe1f9))
* **maintenance:** add admin maintenance dashboard, disk health, risk audit, and recalc-phash CLI ([c0109e8](https://github.com/atbeta/picfast/commit/c0109e89f05f9add3d6382c37b7f6febaba496dd))
* **maintenance:** add database stats, pHash coverage, cleanup tool, and log hints ([bac63d1](https://github.com/atbeta/picfast/commit/bac63d157cfbc921da8e322874cdfe527725ec01))
* **webdav:** support direct 302 redirect for public endpoints ([9e2d19a](https://github.com/atbeta/picfast/commit/9e2d19a53d7d9e7a60d5bbf19fa87122cd4360ba))


### Bug Fixes

* **maintenance:** break recalc-phash infinite loop with cursor pagination ([f861795](https://github.com/atbeta/picfast/commit/f86179507c3ed7b27bdbbcd992480461d2cc7781))
* **service:** cap upload-time pHash memory by direct 32x32 resampling ([10406f5](https://github.com/atbeta/picfast/commit/10406f5bd2baedac4cacf0b9ca14470ddfee6ba2))

## [0.17.1](https://github.com/atbeta/picfast/compare/v0.17.0...v0.17.1) (2026-07-02)


### Bug Fixes

* **web:** suppress zod v4 eval probe to eliminate CSP violation ([f7f0a51](https://github.com/atbeta/picfast/commit/f7f0a516550b8d49e769e0245f5471506561439d))

## [0.17.0](https://github.com/atbeta/picfast/compare/v0.16.0...v0.17.0) (2026-06-28)


### Features

* **docs:** add OpenAPI spec, MCP pipeline tool, and webhook documentation ([eed5b72](https://github.com/atbeta/picfast/commit/eed5b72f3d38098a4e0822a13ded6644ac78fc75))
* **events:** add event system and webhook delivery platform ([a0dce24](https://github.com/atbeta/picfast/commit/a0dce24b795432f198b810f8967adbf3eb9f4be2))
* **pipeline:** add image pipeline status API and detail UI ([1909892](https://github.com/atbeta/picfast/commit/1909892df80328dc836cf5460f70070af2dbcc49))
* **webhook:** add webhook management UI with i18n support ([ee9f450](https://github.com/atbeta/picfast/commit/ee9f450ae1f0e326ca0ecf56f50ed460ae3b0092))


### Bug Fixes

* **pipeline:** add permission check, processing skipped status, i18n, and test ([583141f](https://github.com/atbeta/picfast/commit/583141fc649227af710d710ba0c8d0b342bdf248))

## [0.16.0](https://github.com/atbeta/picfast/compare/v0.15.2...v0.16.0) (2026-06-27)


### Features

* **personalization:** unified preferences with Layer × Type matrix ([8f29b76](https://github.com/atbeta/picfast/commit/8f29b76adfd452eec86833383417ce2c6731a5ba))
* **theme:** support user-scoped theme package import/export ([f36f928](https://github.com/atbeta/picfast/commit/f36f92878495ea869187b69f4d8ad3cb75c98b6e))


### Bug Fixes

* **admin:** stack custom CSS scope hint below the textarea ([7d3eb72](https://github.com/atbeta/picfast/commit/7d3eb7214ad9e36b3002f6cbe3c82a98c3ab2251))
* **personalization:** toast on sync failure, AuthProvider context, i18n keys, rename to Personalization ([00c14ea](https://github.com/atbeta/picfast/commit/00c14eaa0084c7a30c68204f61984174b53ed2e3))
* **ui:** widen image processing dialog to prevent horizontal scroll ([139b51f](https://github.com/atbeta/picfast/commit/139b51f033646f1192f7323e5d5d44564156a1d2))

## [0.15.2](https://github.com/atbeta/picfast/compare/v0.15.1...v0.15.2) (2026-06-22)


### Bug Fixes

* **server:** auto-trust private/loopback IPs for X-Forwarded-For extraction ([92a2c4e](https://github.com/atbeta/picfast/commit/92a2c4e028dd5d4348c7dcfed3615597db11e028)), closes [#55](https://github.com/atbeta/picfast/issues/55)

## [0.15.1](https://github.com/atbeta/picfast/compare/v0.15.0...v0.15.1) (2026-06-21)


### Bug Fixes

* **auth:** SSO login now correctly lands on console after OAuth callback ([7a48dd3](https://github.com/atbeta/picfast/commit/7a48dd3a2790c55183e947a95ff4d1a31f98cbdc)), closes [#70](https://github.com/atbeta/picfast/issues/70)

## [0.15.0](https://github.com/atbeta/picfast/compare/v0.14.1...v0.15.0) (2026-06-19)


### Features

* **admin:** add CSP override fields in analytics settings ([0e776dc](https://github.com/atbeta/picfast/commit/0e776dc16e5487d5108dc0fc2249a14c87ae88e7))


### Bug Fixes

* **csp:** update umami defaults and add config-driven CSP overrides ([51de6a1](https://github.com/atbeta/picfast/commit/51de6a12889847c4b87be7caf53c43f36f21386f))

## [0.14.1](https://github.com/atbeta/picfast/compare/v0.14.0...v0.14.1) (2026-06-19)


### Bug Fixes

* **docker:** replace corepack with npm install to fix arm64 CI build ([32216eb](https://github.com/atbeta/picfast/commit/32216eb86df6f72771b84b3a4570feca742adc4c))

## [0.14.0](https://github.com/atbeta/picfast/compare/v0.13.1...v0.14.0) (2026-06-19)


### Features

* **audit:** add audit logs for album CRUD and moderation review ([c6b33d4](https://github.com/atbeta/picfast/commit/c6b33d4ff22164e94653f1509d89e903284a5ecf))
* **image:** support on-the-fly processing via URL params ([380578e](https://github.com/atbeta/picfast/commit/380578e9ffb0de3f665717060bf097edcf6c2023))
* **strategy:** support link_mode=proxy for cloud storage backends ([7f081f6](https://github.com/atbeta/picfast/commit/7f081f624815745cc59888a4a12bccc90b1314e7))


### Bug Fixes

* **csp:** bundle Scalar locally and tighten connect-src policy ([39057cb](https://github.com/atbeta/picfast/commit/39057cb17aab5d239b40943455d9d61e4ff23fb9))
* **image:** prevent Content-Type mismatch on processing failure ([7884463](https://github.com/atbeta/picfast/commit/7884463a8b019adad38ad67705e2f848e5afea42))
* **upload:** use result.Links in flat and sharex upload handlers ([762a591](https://github.com/atbeta/picfast/commit/762a591f1bf8669e0494b7c0e1aa41be069e3ac6))

## [0.13.1](https://github.com/atbeta/picfast/compare/v0.13.0...v0.13.1) (2026-06-12)


### Bug Fixes

* **web:** resolve auth routing and state loading issues ([0bbc128](https://github.com/atbeta/picfast/commit/0bbc128baed66165d70e2edf3f89bdbc28698cfa))

## [0.13.0](https://github.com/atbeta/picfast/compare/v0.12.1...v0.13.0) (2026-06-09)


### Features

* **group:** add per-group user storage capacity quota ([079f354](https://github.com/atbeta/picfast/commit/079f354ad6935f34167a5c5d0ebf068e5abea41b))


### Bug Fixes

* **server:** unify client IP extraction behind reverse proxies ([4074995](https://github.com/atbeta/picfast/commit/40749953b59c157ab5a8c1238a73fdc3eecdd0d6))

## [0.12.1](https://github.com/atbeta/picfast/compare/v0.12.0...v0.12.1) (2026-06-08)


### Bug Fixes

* **upload:** hide image processing settings when skip processing is enabled ([bfde8e3](https://github.com/atbeta/picfast/commit/bfde8e30b849bb5b4204640b8051c87fda9db1db))

## [0.12.0](https://github.com/atbeta/picfast/compare/v0.11.1...v0.12.0) (2026-06-08)


### Features

* **auth:** separate allow_oauth_registration from allow_registration ([a5a385a](https://github.com/atbeta/picfast/commit/a5a385a5a475a7086406af09cc82e16874ffe4fb))
* **upload:** add skip_image_processing to preserve original images ([79a5ff2](https://github.com/atbeta/picfast/commit/79a5ff2943dbb05c68cc417ba121b9b27ccad2be))


### Bug Fixes

* **middleware:** allow analytics scripts through CSP ([783b415](https://github.com/atbeta/picfast/commit/783b4152fd67026999d1012ad8c19b5bf58aff9c))

## [0.11.1](https://github.com/atbeta/picfast/compare/v0.11.0...v0.11.1) (2026-06-07)


### Bug Fixes

* **oauth:** correct JWKS URL handling and support manual OIDC configuration ([3cbff41](https://github.com/atbeta/picfast/commit/3cbff412cddc16ba85b49dc3b222b1de77a5d165))

## [0.11.0](https://github.com/atbeta/picfast/compare/v0.10.0...v0.11.0) (2026-06-05)


### Features

* **auth:** add OAuth2/OIDC support with PKCE, GitHub provider, and full UI ([c35c795](https://github.com/atbeta/picfast/commit/c35c79546163d7360e1d6076fed42d5cef1ee11f))


### Bug Fixes

* **docker:** set MALLOC_ARENA_MAX=2 to reduce vips memory fragmentation ([9c58d8c](https://github.com/atbeta/picfast/commit/9c58d8cbdbc3816ed41449fa848aebb838c559f3))

## [0.10.0](https://github.com/atbeta/picfast/compare/v0.9.0...v0.10.0) (2026-06-04)


### Features

* **admin:** edit per-minute and per-hour upload limits in groups ([e81eebf](https://github.com/atbeta/picfast/commit/e81eebf04fe364a0f5c11bfe3cd6c2200a25626b))


### Bug Fixes

* **docker:** apply security patches to resolve libgnutls30 CVEs ([232c459](https://github.com/atbeta/picfast/commit/232c459dc6feb609791351dd0c53b354947e9393))

## [0.9.0](https://github.com/atbeta/picfast/compare/v0.8.0...v0.9.0) (2026-06-03)


### Features

* **admin:** add copy templates and batch link export ([e83e23f](https://github.com/atbeta/picfast/commit/e83e23f37091427f93f93b522dbfdd484c8592fc))
* **theme:** add theme packages and semantic preset surfaces ([32d1b3d](https://github.com/atbeta/picfast/commit/32d1b3d41965ee90302473f441242023f7d143e1))


### Bug Fixes

* **security:** harden auth cookies and upload validation ([6abe61c](https://github.com/atbeta/picfast/commit/6abe61c6473a2c4d502f18c87349190b2e8547cf))
* **store:** local storage image URLs redirect to homepage ([#42](https://github.com/atbeta/picfast/issues/42)) ([da1e2ba](https://github.com/atbeta/picfast/commit/da1e2baca87f9ffceb88f159cc770b44b0a57562))

## [0.8.0](https://github.com/atbeta/picfast/compare/v0.7.0...v0.8.0) (2026-05-10)


### Features

* **admin:** customize site appearance with theme presets ([34c9368](https://github.com/atbeta/picfast/commit/34c93680b790c3d00f3b61431d248d7e850bba46))
* **maintenance:** add backup and restore CLI with admin reference ([94ca147](https://github.com/atbeta/picfast/commit/94ca1478a984fb93b7590463458432af1b95aae7))

## [0.7.0](https://github.com/atbeta/picfast/compare/v0.6.4...v0.7.0) (2026-05-05)


### Features

* **storage:** return direct CDN URLs for external storage strategies ([92b2373](https://github.com/atbeta/picfast/commit/92b2373223051ff05dc1126a2e52e6075d2fa3f2))


### Bug Fixes

* **ux:** add execCommand fallback for clipboard copy on older Edge ([25cdea4](https://github.com/atbeta/picfast/commit/25cdea46ace81c0f4e90591aa710305792f62bc2))
* **ux:** improve console mobile navigation and admin readability ([464bd3f](https://github.com/atbeta/picfast/commit/464bd3f1b15345d5db795648934a3fd9bc451c2e))

## [0.6.4](https://github.com/atbeta/picfast/compare/v0.6.3...v0.6.4) (2026-05-05)


### Bug Fixes

* **openapi:** replace hardcoded pbeta.me server URL with localhost placeholder ([c2926ad](https://github.com/atbeta/picfast/commit/c2926ad04032154d7559d21a4f0728d1644d1fc9))

## [0.6.3](https://github.com/atbeta/picfast/compare/v0.6.2...v0.6.3) (2026-05-05)


### Bug Fixes

* **mcp:** unwrap standard API JSON envelope in MCP client ([dd60bb2](https://github.com/atbeta/picfast/commit/dd60bb2a6a1d06fae093f76d44e0e2f805d747b4))
* **upload:** use localStorage-only for permission preference, skip server default ([caf2ec8](https://github.com/atbeta/picfast/commit/caf2ec817b9028b2d3b4c9e2f212bee8b0f8163b))
* **ux:** add cursor-pointer to Button base styles ([517fe67](https://github.com/atbeta/picfast/commit/517fe67cd8c3383677b626f1c8472085e010c872))


### Performance Improvements

* **docker:** replace 90MB fonts-noto-cjk with 5MB fonts-wqy-microhei ([d7c50d0](https://github.com/atbeta/picfast/commit/d7c50d08a2296e8454ec443ed1025c740e7a58db))

## [0.6.2](https://github.com/atbeta/picfast/compare/v0.6.1...v0.6.2) (2026-05-04)


### Bug Fixes

* **admin:** populate user count in group list endpoint ([5af02e9](https://github.com/atbeta/picfast/commit/5af02e95135f359dc6a54f9d19fc592ddbb84da4))
* **upload:** persist selected permission across page navigation ([0f63f01](https://github.com/atbeta/picfast/commit/0f63f01e1fbe43eafba46ab5f4bb29558d683971))
* **upload:** update extension when image format is converted during processing ([be4998c](https://github.com/atbeta/picfast/commit/be4998c608cf39f68908ff9e5b76035a473ef108))
* **ux:** render SVG/ICO originals as thumbnails when no preview is generated ([dba2f27](https://github.com/atbeta/picfast/commit/dba2f27e50ebdb9006b8db0d99bc3046a66618de))

## [0.6.1](https://github.com/atbeta/picfast/compare/v0.6.0...v0.6.1) (2026-05-04)


### Bug Fixes

* **admin-settings:** restore image lifecycle dropdown value on page reload ([da78a1b](https://github.com/atbeta/picfast/commit/da78a1b595d21799acc6dbca7ef839e9000f428b))
* **admin:** fallback to uploaded IP when image has no associated user email ([bdc84fa](https://github.com/atbeta/picfast/commit/bdc84fa0fd8bdc2318cb4e4d4c9b3f146f758e22))
* **thumbnail:** skip thumbnail URL for SVG and ICO formats ([f8a0ded](https://github.com/atbeta/picfast/commit/f8a0dedc5026b6a573839b8174f83cb10635c9c5))
* **ux:** keep previous image list data while search is in flight ([072fc1c](https://github.com/atbeta/picfast/commit/072fc1ca3e4c3d3fed8ef224a12b4c16bf174ff8))

## [0.6.0](https://github.com/atbeta/picfast/compare/v0.5.0...v0.6.0) (2026-05-04)


### Features

* **auth:** enhance registration flow and add password reset ([ed44113](https://github.com/atbeta/picfast/commit/ed44113ba23ffff62ac4c59db041c5bcfee66281))
* **images:** add batch delete endpoint to reduce N requests to 1 ([7618aba](https://github.com/atbeta/picfast/commit/7618aba1d6ae85c690e4b78908867508059d7e76))
* **images:** add keyword, extension, and date-range filtering for image lists ([05fd1b9](https://github.com/atbeta/picfast/commit/05fd1b9d272c94c0f4d46cbd90d471fb380f8b7c))
* **observability:** expose Docker-ready metrics ([9a4f0f7](https://github.com/atbeta/picfast/commit/9a4f0f7f2ba391f0119672ccdfa37050fd20cbc6))


### Bug Fixes

* **auth:** harden email verification and reset mail flow ([5b26f6b](https://github.com/atbeta/picfast/commit/5b26f6b27fc9e591aaeafbd7ab3ae53a51bea3ec))
* **config:** load local .env and align local dev defaults ([e096391](https://github.com/atbeta/picfast/commit/e096391ba76b958c54a09acd72b5ec97c2535c28))
* **ux:** make toast notifications auto-dismiss and unify save feedback ([56392b7](https://github.com/atbeta/picfast/commit/56392b7a1eaa9abaf0866b5d3327e64033323216))

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
