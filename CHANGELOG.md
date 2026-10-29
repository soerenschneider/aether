# Changelog

## [1.2.0](https://github.com/soerenschneider/aether/compare/v1.1.1...v1.2.0) (2026-10-29)


### Features

* add ability to display victorialogs logs ([7d17d4b](https://github.com/soerenschneider/aether/commit/7d17d4be9826ba60dce706da6c06d6119d65dda3))
* **cache:** show outdated data when updates fail ([f624773](https://github.com/soerenschneider/aether/commit/f6247730bb471e0abec92318bd9a5248d0dbb5b9))
* **caldav:** expand recurring events ([21a69ee](https://github.com/soerenschneider/aether/commit/21a69ee8cfe3912a035e83aadd99c32a9b23c413))
* **caldav:** support multiple calendars ([1ff912c](https://github.com/soerenschneider/aether/commit/1ff912c026d4758fa8e192319ed80af2542d49be))
* **calendar:** add month view ([7e0b4d5](https://github.com/soerenschneider/aether/commit/7e0b4d5bfd5b9fc1df3f04b53796456283176f28))
* **calendar:** group agenda by day and show today's anniversaries ([5829324](https://github.com/soerenschneider/aether/commit/5829324faff99e97feb55e5c3fbf085fd620017d))
* **dev:** add dev server with fake data ([b1e3814](https://github.com/soerenschneider/aether/commit/b1e3814d0d6535e819d171159765d93507b03596))
* **email:** send an email-friendly version of the dashboard ([580c037](https://github.com/soerenschneider/aether/commit/580c0372db8991562c841c07ef803a4fa23d84ae))
* **tasks:** group tasks by due date ([dce0804](https://github.com/soerenschneider/aether/commit/dce080458a652e1426142a259e39592314302155))
* **ui:** pack cards and cap long lists ([d5e5978](https://github.com/soerenschneider/aether/commit/d5e59782f4b426d8e1fb307e70ef1e4c85a74235))
* **ui:** redesign dashboard and refresh it with htmx ([1cd9c2d](https://github.com/soerenschneider/aether/commit/1cd9c2db1a32ead4bc2e0263cb0d35cbd0289591))
* **ui:** redesign dashboard and refresh it with htmx ([05c8cf9](https://github.com/soerenschneider/aether/commit/05c8cf92807cc309dd637d6f62ca6f2ffe334995))


### Bug Fixes

* address golangci-lint findings ([77d53ce](https://github.com/soerenschneider/aether/commit/77d53ceeead7b3f52c8476ca3dc84b5185ad2c98))
* **cache:** retry incomplete data sooner ([a96a775](https://github.com/soerenschneider/aether/commit/a96a7751bbae0f89f8d66bafdb24e47711494f4a))
* **caldav:** show all-day events of today ([9f98681](https://github.com/soerenschneider/aether/commit/9f986812c0d89007b0d5b359d97d6d07f69da91a))
* **carddav:** count age towards the next anniversary ([acb81ce](https://github.com/soerenschneider/aether/commit/acb81ce5c1779e3968fe81809f7ee7625539482c))
* count time per datasource ([83cab32](https://github.com/soerenschneider/aether/commit/83cab32dd7b35aad35a05e12cfee4b9d099beb7a))
* **deps:** bump alpine from 3.21.2 to 3.21.3 ([7aa367a](https://github.com/soerenschneider/aether/commit/7aa367a962e93936da17132d96e97cf1e7d7d749))
* **deps:** bump alpine from 3.21.2 to 3.21.3 ([c5af587](https://github.com/soerenschneider/aether/commit/c5af587e887960143fd2bbcf94171828cdac603a))
* **deps:** bump github.com/go-playground/validator/v10 ([41157ef](https://github.com/soerenschneider/aether/commit/41157ef037dcac9a2c15ac1bdd1f103ff7ab68ab))
* **deps:** bump github.com/go-playground/validator/v10 from 10.24.0 to 10.25.0 ([df72a6e](https://github.com/soerenschneider/aether/commit/df72a6e5e03e678667afe9dd301a2f49d360a605))
* **deps:** bump github.com/prometheus/alertmanager from 0.28.0 to 0.28.1 ([5278910](https://github.com/soerenschneider/aether/commit/5278910bc333bc2318c59c5f4ecd28bbb76dee9d))
* **deps:** bump github.com/prometheus/alertmanager from 0.28.0 to 0.28.1 ([c07df28](https://github.com/soerenschneider/aether/commit/c07df28d88c99cd516678bf1f67bb1cfd190c893))
* **deps:** bump github.com/prometheus/alertmanager from 0.28.0 to 0.28.1 ([1040a7e](https://github.com/soerenschneider/aether/commit/1040a7eb76bb072278f557a635b880b3533922ca))
* **deps:** bump github.com/prometheus/client_golang ([bad9ce2](https://github.com/soerenschneider/aether/commit/bad9ce26304281139aa3fd968d7b6b7f0edc2b43))
* **deps:** bump github.com/prometheus/client_golang from 1.20.5 to 1.21.1 ([c07266d](https://github.com/soerenschneider/aether/commit/c07266df6699ea0c3b58ef3732c7b3a2fe2a1eb9))
* **deps:** bump golang from 1.23.6 to 1.24.1 ([a774b2e](https://github.com/soerenschneider/aether/commit/a774b2e924829ff91632fcad83f23d3692f2d0f4))
* **deps:** bump golang from 1.23.6 to 1.24.1 ([a9e8079](https://github.com/soerenschneider/aether/commit/a9e807917ef376469623ff60c50b0f223a99a992))
* **deps:** bump golang.org/x/net from 0.34.0 to 0.36.0 ([b0c1042](https://github.com/soerenschneider/aether/commit/b0c10420a24b13bbc87cb41019e481eb037e1d19))
* **deps:** bump golang.org/x/net from 0.34.0 to 0.36.0 ([ea36a95](https://github.com/soerenschneider/aether/commit/ea36a9595814dea56646b588138ae945afcddad9))
* don't fail when single datasource returns error ([1c1fa20](https://github.com/soerenschneider/aether/commit/1c1fa200d822d546af33ba1b21e603b5598039bd))
* pluralize durations correctly ([5715e45](https://github.com/soerenschneider/aether/commit/5715e4520beabf1c50ff317a0c67afab0823c6a8))
* strip trailing newlines from secret files ([a4967a1](https://github.com/soerenschneider/aether/commit/a4967a1109869477460af436060f2c42e6e2ef97))
* use custom template files instead of the default templates ([f41b63b](https://github.com/soerenschneider/aether/commit/f41b63b308e03476cd0d73f20ff5a61fab2db625))

## [1.1.1](https://github.com/soerenschneider/aether/compare/v1.1.0...v1.1.1) (2025-02-17)


### Bug Fixes

* improve mobile css for better readability ([14cb27f](https://github.com/soerenschneider/aether/commit/14cb27f831c871c44a48e3761b4f372468471b50))
* show correct number of days ([88b4e05](https://github.com/soerenschneider/aether/commit/88b4e0514121b0aab0fe6ba1057135c69f57d6e9))

## [1.1.0](https://github.com/soerenschneider/aether/compare/v1.0.0...v1.1.0) (2025-02-11)


### Features

* configure lookahead days for taskwarrior summary ([9c75cf4](https://github.com/soerenschneider/aether/commit/9c75cf44eea7e56af2b39e953878c20292f6a2bb))


### Bug Fixes

* fix calculating even/odd classes ([3d2b828](https://github.com/soerenschneider/aether/commit/3d2b82833fc1e4266e13c62a41af2d1a96b27936))
* fix css summary table rendering ([42d2f9c](https://github.com/soerenschneider/aether/commit/42d2f9cd24535b827f7cce8cc00fc5f2a0299735))
* fix font size when displaying specific emojis ([daca2d3](https://github.com/soerenschneider/aether/commit/daca2d33a761910b1a0a5e2fbd4c6289612604a4))

## 1.0.0 (2025-02-10)


### Features

* add ids to templates to make links work ([96fb4bd](https://github.com/soerenschneider/aether/commit/96fb4bdec84c3b2e9394fcd129b12402bc4f7b58))
* allow exclusion of datasources from summary ([4aada91](https://github.com/soerenschneider/aether/commit/4aada91be6fd91acf802ecbe4b771115e9ff14ca))
* improve templates ([98f5053](https://github.com/soerenschneider/aether/commit/98f5053a3a551d343bdf80a6671fa6c4551b858c))


### Bug Fixes

* fix minifier ([db5bccb](https://github.com/soerenschneider/aether/commit/db5bccb18d34379f27541912171e62c5b2cdc938))
* fix parameter mixup ([c713165](https://github.com/soerenschneider/aether/commit/c7131651545f48880dff3362b8f5250a1da9c896))

## [1.0.2](https://github.com/soerenschneider/aether/compare/v1.0.1...v1.0.2) (2025-02-02)


### Bug Fixes

* **deps:** bump alpine from 3.19.1 to 3.21.2 ([bf9712c](https://github.com/soerenschneider/aether/commit/bf9712ca3903c51e501b43cb3813ef16348a6580))
* **deps:** bump github.com/rs/zerolog from 1.32.0 to 1.33.0 ([3a306ff](https://github.com/soerenschneider/aether/commit/3a306ff4cfd9c5f7583f2313f2a0100ea1843e4b))
* **deps:** bump github.com/sj14/astral from 0.2.0 to 0.2.2 ([5a71a7d](https://github.com/soerenschneider/aether/commit/5a71a7dfff0ad8f0cb8323e9e3873a0ca5e27356))
* **deps:** bump golang from 1.22.2 to 1.23.5 ([6646149](https://github.com/soerenschneider/aether/commit/6646149a505de50bb61ef795c59b6b9d948d5643))
* fix compliation errors ([85c2808](https://github.com/soerenschneider/aether/commit/85c2808eaffc5d6ab14661a073f3052b3251c2c8))

## [1.0.1](https://github.com/soerenschneider/aether/compare/v1.0.0...v1.0.1) (2025-02-02)


### Bug Fixes

* **deps:** bump github.com/emersion/go-webdav from 0.4.0 to 0.6.0 ([4176ca5](https://github.com/soerenschneider/aether/commit/4176ca5b62781e55aaee68325c720babcbc7017b))
* **deps:** bump github.com/go-playground/validator/v10 ([7ef20ee](https://github.com/soerenschneider/aether/commit/7ef20ee64bff89d4553ebd61cd99b299bba7133c))
* **deps:** bump github.com/hashicorp/go-retryablehttp ([2f4cf91](https://github.com/soerenschneider/aether/commit/2f4cf91fe667ddbde2eb5ee4dd519794684b3dff))
* **deps:** bump github.com/prometheus/alertmanager from 0.26.0 to 0.28.0 ([20efe67](https://github.com/soerenschneider/aether/commit/20efe671347e7ce25293e30c384bd4f1c6feb920))
* **deps:** bump golang from 1.22.0 to 1.22.2 ([a71bd11](https://github.com/soerenschneider/aether/commit/a71bd1127d5d91634de9b6964630645c0014dfda))
* **deps:** bump golang.org/x/crypto from 0.19.0 to 0.31.0 ([308e22a](https://github.com/soerenschneider/aether/commit/308e22a06e1559721226317c2384d0aadf0f222e))
* fix email reports ([0a27d6f](https://github.com/soerenschneider/aether/commit/0a27d6f4afe11082f001c937ff38bdf23c447599))
* fix potential race condition ([9abe28c](https://github.com/soerenschneider/aether/commit/9abe28cc7e7590d8ba041f837935c1ff2b5e5888))

## 1.0.0 (2024-02-11)


### Bug Fixes

* **deps:** bump github.com/go-co-op/gocron from 1.36.0 to 1.37.0 ([d252002](https://github.com/soerenschneider/aether/commit/d2520024cd28022091afe1508537e47c04914da4))
* **deps:** bump github.com/go-openapi/strfmt from 0.21.7 to 0.22.0 ([e34248f](https://github.com/soerenschneider/aether/commit/e34248f86c8886e6492460ed473ea9b3345b4398))
* **deps:** bump github.com/go-playground/validator/v10 ([c81f70f](https://github.com/soerenschneider/aether/commit/c81f70f2956462b118d0c8d46d4c0e23237c1e16))
* **deps:** bump github.com/rs/zerolog from 1.31.0 to 1.32.0 ([b5ca7d8](https://github.com/soerenschneider/aether/commit/b5ca7d849482ba1895f85a656a206071edb1604f))

## [1.1.0](https://github.com/soerenschneider/aether/compare/v1.0.0...v1.1.0) (2023-11-20)


### Features

* add container support ([cf0e878](https://github.com/soerenschneider/aether/commit/cf0e878b06b618f9ceaca7a01a937f6dd555675b))
* support env var support ([3357e56](https://github.com/soerenschneider/aether/commit/3357e5697c8d0e21caf69c5a60ee48c6663ff284))

## 1.0.0 (2023-11-18)


### Bug Fixes

* **deps:** bump github.com/go-co-op/gocron from 1.35.3 to 1.36.0 ([7f6be6a](https://github.com/soerenschneider/aether/commit/7f6be6a48221ec674c85dbea5f16e7909e2512c1))
* supply opts to taskwarrior datasource ([c9ac70e](https://github.com/soerenschneider/aether/commit/c9ac70eecede579e84df3c87b1f0df77b120d585))
