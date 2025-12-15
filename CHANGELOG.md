# [1.10.0](https://github.com/vanducng/cflip/compare/v1.9.0...v1.10.0) (2025-12-15)


### Features

* add list command to show providers and current selection ([82170ce](https://github.com/vanducng/cflip/commit/82170ce76561b07d89ce71cfb07c671d2eec6df0))
* remove API key prompt for Anthropic subscription plan ([a9b01d7](https://github.com/vanducng/cflip/commit/a9b01d78af92df6f366f71c741f94a43062357cb))

# [1.9.0](https://github.com/vanducng/cflip/compare/v1.8.0...v1.9.0) (2025-12-15)


### Features

* enhance snapshots and provider display ([efdb26f](https://github.com/vanducng/cflip/commit/efdb26f3fe2b6849a43ac6de37565b367b877ff2))

# [1.8.0](https://github.com/vanducng/cflip/compare/v1.7.0...v1.8.0) (2025-12-15)


### Features

* always create snapshot before switching providers ([141324d](https://github.com/vanducng/cflip/commit/141324df4181537c5a9df2735875b3838b999bf8))

# [1.7.0](https://github.com/vanducng/cflip/compare/v1.6.8...v1.7.0) (2025-12-15)


### Features

* comprehensive settings management with snapshots ([5489b0f](https://github.com/vanducng/cflip/commit/5489b0f28f0dbef6169ee3c75b418f4d5c54018e))

## [1.6.8](https://github.com/vanducng/cflip/compare/v1.6.7...v1.6.8) (2025-12-15)


### Bug Fixes

* set version on root command at runtime to ensure version is displayed ([c4cb650](https://github.com/vanducng/cflip/commit/c4cb6500f813576254fbd348a314ab19273eb9f0))

## [1.6.7](https://github.com/vanducng/cflip/compare/v1.6.6...v1.6.7) (2025-12-15)


### Bug Fixes

* preserve Anthropic API key when switching providers ([d126802](https://github.com/vanducng/cflip/commit/d12680295e90bac654ed5a871c568f9b6cb533dc))

## [1.6.6](https://github.com/vanducng/cflip/compare/v1.6.5...v1.6.6) (2025-12-15)


### Bug Fixes

* add missing version variable for goreleaser ldflags ([9a83715](https://github.com/vanducng/cflip/commit/9a837159b9378cb5e71046655756555a0bee8fd0))

## [1.6.5](https://github.com/vanducng/cflip/compare/v1.6.4...v1.6.5) (2025-12-15)


### Bug Fixes

* use Formula directory for homebrew tap instead of Casks ([8b2b335](https://github.com/vanducng/cflip/commit/8b2b335c2999b2f36769b983c0d2f724f0380399))

## [1.6.4](https://github.com/vanducng/cflip/compare/v1.6.3...v1.6.4) (2025-12-15)


### Bug Fixes

* update version example in README to recent release ([db96038](https://github.com/vanducng/cflip/commit/db96038b962fae9258eb625bd5f3997753ca19f9))

## [1.6.3](https://github.com/vanducng/cflip/compare/v1.6.2...v1.6.3) (2025-12-15)


### Bug Fixes

* pass HOMEBREW_TAP_GITHUB_TOKEN to goreleaser ([96e6f83](https://github.com/vanducng/cflip/commit/96e6f83fbf5d56d72849c1d2af4ed438206ebeb5))
* update goreleaser config to fix deprecation warnings ([0b2c771](https://github.com/vanducng/cflip/commit/0b2c77120c462739625e5d6dce8734348925067d))

## [1.6.2](https://github.com/vanducng/cflip/compare/v1.6.1...v1.6.2) (2025-12-15)


### Bug Fixes

* clean npm artifacts before goreleaser to avoid dirty git state ([e09dbf2](https://github.com/vanducng/cflip/commit/e09dbf23a36866519bcab55f8dc6e30e690b14e3))

## [1.6.1](https://github.com/vanducng/cflip/compare/v1.6.0...v1.6.1) (2025-12-15)


### Bug Fixes

* combine semantic-release and goreleaser in single workflow ([3686cef](https://github.com/vanducng/cflip/commit/3686cefb988ae583bc85224bf6162d3568f02a3f))

# [1.6.0](https://github.com/vanducng/cflip/compare/v1.5.2...v1.6.0) (2025-12-15)


### Bug Fixes

* correct release detection regex to match commit subjects ([3104443](https://github.com/vanducng/cflip/commit/3104443de08c945d372bf1cecdd360fc987620c5))
* resolve version assignment and formatting issues ([d4b16eb](https://github.com/vanducng/cflip/commit/d4b16ebaf122e2f77a3342bcd83ecdb6be35b79f))
* resolve version redeclaration and add pre-commit checks ([58dba51](https://github.com/vanducng/cflip/commit/58dba510defabd96ea7b70f71155366fb3d757c8))


### Features

* add version command and display ([681ae00](https://github.com/vanducng/cflip/commit/681ae00b13983e2216a4a36e73a77225e585450f))
* improve cache configuration and fix release pipeline ([50f5d8a](https://github.com/vanducng/cflip/commit/50f5d8ad6ee441ca4b6d66485f52a517a20b8c54))
* refactor the release pipeline ([0e384af](https://github.com/vanducng/cflip/commit/0e384afd5d9c5fe13583b01db7107a88e206ae1e))

## [1.5.2](https://github.com/vanducng/cflip/compare/v1.5.1...v1.5.2) (2025-12-15)


### Bug Fixes

* restructure CI/CD workflows to follow proper sequence ([3c9f292](https://github.com/vanducng/cflip/commit/3c9f2929a661e9a37a60e5ba466e1cc1e2c865bb))

## [1.5.1](https://github.com/vanducng/cflip/compare/v1.5.0...v1.5.1) (2025-12-15)


### Bug Fixes

* improve API key persistence and fix CI cache error ([7ddc231](https://github.com/vanducng/cflip/commit/7ddc23194674928816e1dc3788cb9e33e2c89869))

# [1.5.0](https://github.com/vanducng/cflip/compare/v1.4.1...v1.5.0) (2025-12-15)


### Features

* update to latest 2025 model versions ([040b958](https://github.com/vanducng/cflip/commit/040b9580b0c4e19c95313f9819489b95d58ac99f))

## [1.4.1](https://github.com/vanducng/cflip/compare/v1.4.0...v1.4.1) (2025-12-15)


### Bug Fixes

* update Anthropic models with latest capabilities ([8d0235f](https://github.com/vanducng/cflip/commit/8d0235f2c9df7832c57aba0c0e2e60b3952065a6))

# [1.4.0](https://github.com/vanducng/cflip/compare/v1.3.2...v1.4.0) (2025-12-14)


### Features

* add support for Claude Code subscription authentication and centralized config ([4c28b0b](https://github.com/vanducng/cflip/commit/4c28b0bc19b4d59bd8ddc37a96228486ec626257))

## [1.3.1](https://github.com/vanducng/cflip/compare/v1.3.0...v1.3.1) (2025-12-14)


### Bug Fixes

* **ci:** configure Homebrew tap token for formula publishing ([622ccaa](https://github.com/vanducng/cflip/commit/622ccaa2ff231ff6f2092f2d1e872cce5884f4a4))

# [1.3.0](https://github.com/vanducng/cflip/compare/v1.2.0...v1.3.0) (2025-12-14)


### Features

* **ci:** integrate Homebrew publishing with GoReleaser ([d4b519e](https://github.com/vanducng/cflip/commit/d4b519e13a1aa6c9f2fb18729dd9b76b369f19a8))

# [1.2.0](https://github.com/vanducng/cflip/compare/v1.1.2...v1.2.0) (2025-12-14)


### Features

* **release:** update GoReleaser and install script ([7d97c99](https://github.com/vanducng/cflip/commit/7d97c99b747f35389a02a2831eaf53798312f804))

## [1.1.2](https://github.com/vanducng/cflip/compare/v1.1.1...v1.1.2) (2025-12-14)


### Bug Fixes

* **ci:** resolve cache conflicts and codecov warnings ([53526ab](https://github.com/vanducng/cflip/commit/53526abaf5aaa0e8b30520c75282837a659f57d9))

## [1.1.1](https://github.com/vanducng/cflip/compare/v1.1.0...v1.1.1) (2025-12-14)


### Bug Fixes

* **release:** trigger GoReleaser on main branch ([837ece5](https://github.com/vanducng/cflip/commit/837ece5f26dda134abb0ed8e2bf2088b537ef2dc))

# [1.1.0](https://github.com/vanducng/cflip/compare/v1.0.0...v1.1.0) (2025-12-14)


### Features

* **ci:** add semantic release automation ([d0d132d](https://github.com/vanducng/cflip/commit/d0d132dca8c76562542999f5c621a404991eb32a))
