# NineSong Public Snapshot

Language: [English](README.md) | [Chinese](README.zh-CN.md)

NineSong is the server-side foundation behind the NSMusicS ecosystem. In its current public form, this repository serves as a frozen open-source snapshot for the music-scene backend, while the broader product direction continues toward a personal digital center spanning music, gallery, video, notes, novels, lists, resource management, and more.

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Super-Badmen-Viper/NineSong)

## Current Public Status And Ecosystem Entry

As of March 22, 2026:

- This public repository is intentionally frozen at its current commit.
- DockerHub distribution remains free for public users.
- The archived public roadmap described React-based frontend apps for music, gallery, video, notes, novels, lists, resource management, and related scenarios as the next open-source application wave, expected to begin around April 2026.
- The paired NSMusicS Windows client is already live on Microsoft Store.
- Direct Microsoft Store link: [ms-windows-store://pdp/?productid=9N0RWS2TJXG1](ms-windows-store://pdp/?productid=9N0RWS2TJXG1)
- A 15-day free trial is currently available on the Windows client.
- App Store releases for macOS and iOS, plus the Google Play release for Android, remain part of the broader client rollout after the current Windows track.
- Based on the current ecosystem planning, the refactored NSMusicS client edition is expected around two months later, which currently points to roughly May 2026.

If you want the newest production-facing client today, start from Microsoft Store.  
If you want the public backend architecture, deployment references, API surface, and archived capability scope, this repository remains the public server-side reference.

## Why NineSong

NineSong is not positioned as a minimal media server.

- It focuses on richer music-library organization instead of only basic streaming.
- It emphasizes metadata, tags, search behavior, CUE-related workflows, and recommendation potential.
- It is designed as a cloud-native backend for a broader personal digital center rather than a single-purpose music product.
- It pairs with NSMusicS while also pointing toward gallery, video, notes, novels, lists, resource management, documents, and related multi-scenario data workflows.

Core stack in this snapshot:

- Go
- Gin
- MongoDB
- JWT-based authentication
- Docker

## Archived Public Capability Snapshot

The archived public README and the codebase describe these public-facing strengths:

- Rich music-library sorting, filtering, and search flows
- Deeper composite-tag processing and metadata-driven organization
- Fuzzy retrieval across title, album, artist, and lyrics
- Support for Chinese pinyin and simplified or traditional Chinese matching
- CUE-oriented playback and file-management workflows
- Music recommendations based on user data and metadata
- A broader architecture aimed at future multimedia and personal digital center scenarios

For the full archived feature scope and the original roadmap-oriented detail, see:

- [doc/Public_Status_Notice.md](doc/Public_Status_Notice.md)
- [doc/Public_Status_Notice.zh-CN.md](doc/Public_Status_Notice.zh-CN.md)
- [doc/Archived_Feature_And_Roadmap.md](doc/Archived_Feature_And_Roadmap.md)
- [doc/Archived_Feature_And_Roadmap.zh-CN.md](doc/Archived_Feature_And_Roadmap.zh-CN.md)

## Deployment And API References

For the historical deployment flow, local debugging notes, and public API references, use:

- [doc/Deployment_And_Local_Development.md](doc/Deployment_And_Local_Development.md)
- [doc/Deployment_And_Local_Development.zh-CN.md](doc/Deployment_And_Local_Development.zh-CN.md)
- [NineSong/api/docs/README.md](NineSong/api/docs/README.md)
- [NineSong/api/docs/API_Documentation.md](NineSong/api/docs/API_Documentation.md)
- [NineSong/api/docs/client_examples/QuickStartGuide.md](NineSong/api/docs/client_examples/QuickStartGuide.md)
- [NineSong API.postman_collection.json](NineSong%20API.postman_collection.json)

## Ecosystem Links

- NSMusicS repository: https://github.com/Super-Badmen-Viper/NSMusicS
- NineSong releases: https://github.com/Super-Badmen-Viper/NineSong/releases/
- Microsoft Store client: [ms-windows-store://pdp/?productid=9N0RWS2TJXG1](ms-windows-store://pdp/?productid=9N0RWS2TJXG1)

## Community

- QQ Group 1: full
- QQ Group 2: `610551734`

## Related Projects And Thanks

- [go-backend-clean-architecture](https://github.com/amitshekhariitbhu/go-backend-clean-architecture)
- [go-audio](https://github.com/go-audio)
- [go-taglib](https://github.com/sentriz/go-taglib)

## Vision

This public repository is frozen, but the broader ecosystem direction continues.  
The long-term goal remains a full personal digital center across multiple clients, stores, and deployment models, with more application scenarios expected to roll out over time, including but not limited to gallery, video, notes, novels, lists, resource management, and more.

The name NineSong comes from "Nine Song" and is inspired by *Chu Ci* and the legacy of Qu Yuan.
