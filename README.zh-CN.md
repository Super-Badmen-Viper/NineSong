# NineSong 公开快照

语言版本：[English](README.md) | [中文](README.zh-CN.md)

NineSong 是 NSMusicS 生态里的服务端基础项目。以当前公开仓库形态来看，这个仓库更适合作为音乐场景后端的冻结开源快照，而更广义的产品方向仍然会继续沿着重构、客户端扩展和后续前端应用推进。

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Super-Badmen-Viper/NineSong)

## 当前公开状态与生态入口

截至 2026 年 3 月 22 日：

- 当前公开仓库有意停留在现有提交，不再作为持续更新的开源后端主仓库。
- DockerHub 分发仍然保持免费。
- 按归档下来的公开路线图，面向音乐、相册、影视、笔记等场景的 React 前端应用，是下一波开源应用方向，预计会在 2026 年 4 月左右开始陆续推出。
- 与 NineSong 配套的 NSMusicS Windows 客户端已经上架微软商城。
- 微软商城直达链接：[ms-windows-store://pdp/?productid=9N0RWS2TJXG1](ms-windows-store://pdp/?productid=9N0RWS2TJXG1)
- Windows 客户端当前提供 15 天免费试用。
- macOS / iOS 的 App Store 版本，以及 Android 的 Google Play 版本，也仍然在当前 Windows 节奏之后的后续发布规划中。
- 按当前生态规划，重构版 NSMusicS 客户端预计约两个月后推出，按现在节奏大致对应 2026 年 5 月。

如果你现在就想体验面向用户的最新客户端，优先从微软商城开始。  
如果你想查看公开的后端架构、部署说明、API 范围和历史能力边界，这个仓库仍然是公开服务端参考入口。

## 为什么是 NineSong

NineSong 的定位并不是一个只负责基础流媒体播放的轻量服务端。

- 更强调完整的音乐库整理能力，而不只是基础串流。
- 更强调元数据、标签、搜索行为、CUE 工作流和推荐潜力。
- 更接近一个可扩展的云原生后端，能从音乐场景继续延伸到更完整的个人数字中心方向。
- 它既服务于 NSMusicS，也面向相册、影视、笔记、文档等多场景数据工作流的长期演进。

当前公开快照的核心技术栈包括：

- Go
- Gin
- MongoDB
- 基于 JWT 的认证
- Docker

## 历史公开能力快照

从归档下来的公开 README 和当前代码结构来看，这个公开版本主要体现出的方向包括：

- 更丰富的音乐库排序、筛选和搜索体验
- 更深入的复合标签与元数据整理能力
- 面向标题、专辑、艺术家、歌词的模糊检索
- 对中文拼音、简繁中文混合匹配的支持
- 面向 CUE 的播放与文件管理工作流
- 基于用户数据与元数据的音乐推荐能力
- 面向更完整多媒体与个人数字中心的长期架构方向

如果你想查看归档下来的完整公开说明和原始路线图细节，请继续看这些文档：

- [doc/Public_Status_Notice.md](doc/Public_Status_Notice.md)
- [doc/Public_Status_Notice.zh-CN.md](doc/Public_Status_Notice.zh-CN.md)
- [doc/Archived_Feature_And_Roadmap.md](doc/Archived_Feature_And_Roadmap.md)
- [doc/Archived_Feature_And_Roadmap.zh-CN.md](doc/Archived_Feature_And_Roadmap.zh-CN.md)

## 部署与 API 入口

如果你要查看历史部署流程、本地调试说明和公开 API 资料，请直接看：

- [doc/Deployment_And_Local_Development.md](doc/Deployment_And_Local_Development.md)
- [doc/Deployment_And_Local_Development.zh-CN.md](doc/Deployment_And_Local_Development.zh-CN.md)
- [NineSong/api/docs/README.md](NineSong/api/docs/README.md)
- [NineSong/api/docs/API_Documentation.md](NineSong/api/docs/API_Documentation.md)
- [NineSong/api/docs/client_examples/QuickStartGuide.md](NineSong/api/docs/client_examples/QuickStartGuide.md)
- [NineSong API.postman_collection.json](NineSong%20API.postman_collection.json)

## 生态入口

- NSMusicS 仓库：https://github.com/Super-Badmen-Viper/NSMusicS
- NineSong Releases：https://github.com/Super-Badmen-Viper/NineSong/releases/
- 微软商城客户端：[ms-windows-store://pdp/?productid=9N0RWS2TJXG1](ms-windows-store://pdp/?productid=9N0RWS2TJXG1)

## 社区

- QQ 群 1：已满
- QQ 群 2：`610551734`

## 相关项目与致谢

- [go-backend-clean-architecture](https://github.com/amitshekhariitbhu/go-backend-clean-architecture)
- [go-audio](https://github.com/go-audio)
- [go-taglib](https://github.com/sentriz/go-taglib)

## 愿景

这个公开仓库已经冻结，但整个生态方向仍然在继续。  
长期目标仍然是从音乐优先的平台出发，继续延展到多客户端、多商店、多部署形态下的更完整个人数字中心体系。

NineSong 的名称来自 “Nine Song | 九歌”，灵感来源于《楚辞》与屈原。
