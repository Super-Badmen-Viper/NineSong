# 部署与本地开发说明

这个文档用于保留旧版公开 README 里的部署和开发说明，同时把文件名对齐到当前仓库里的实际结构。

## 公开 Docker 部署

旧版公开说明里提到的发布包下载入口是：

- https://github.com/Super-Badmen-Viper/NineSong/releases/

归档下来的部署流程大致是：

1. 先下载发布包。
2. 将 `.env` 和 `docker-compose.yaml` 放在同一目录。
3. 按需要调整环境变量和卷映射，尤其是媒体库的宿主机路径。
4. 运行：

```sh
docker compose up -d
```

## 当前公开快照里的默认初始化账户

根据 `NineSong/bootstrap/init.go` 里的初始化代码，当前公开快照中能看到的默认初始化账户为：

- 登录邮箱：`admin@gmail.com`
- 登录密码：`admin123`

## 历史说明里的重要提醒

- 更新镜像版 NineSong 时，媒体库里的临时生成资源，例如专辑封面，可能会被清除。
- 这类临时资源通常需要通过重新扫描媒体库来恢复。
- 如果要彻底重装，需要同时清理容器和 Docker 卷数据。

## 本地开发参考

当前仓库里可直接参考的本地相关文件包括：

- Windows 本地初始化脚本：`NineSong/docker-compose-local-develop-init-windows.ps1`
- 本地 MongoDB Compose 文件：`NineSong/docker-compose-local-mongo.yaml`
- 主 Compose 文件：`NineSong/docker-compose.yaml`

旧版 README 里建议的本地环境调整包括：

- `DB_HOST=localhost`
- `DB_PORT=27017`

当前仓库中的 Windows 初始化脚本会准备这些默认卷路径：

- `C:\Users\Public\Documents\NineSong\MongoDB`
- `C:\Users\Public\Documents\NineSong\Sqlite`
- `C:\Users\Public\Documents\NineSong\MetaData`
- `E:\0_Music`

旧版 README 里还提到，本地实时开发可以使用：

```sh
go install github.com/air-verse/air@latest
air
```

## 仓库里已经自带的 API 与客户端文档

- 索引：`NineSong/api/docs/README.md`
- 完整 API 文档：`NineSong/api/docs/API_Documentation.md`
- 功能总结：`NineSong/api/docs/FeatureSummary.md`
- 快速入门：`NineSong/api/docs/client_examples/QuickStartGuide.md`
- 客户端实现指南：`NineSong/api/docs/client_examples/MediaLibraryClientImplementationGuide.md`
- API 使用示例：`NineSong/api/docs/client_examples/MediaLibraryAPIUsageExamples.md`
- 同步机制说明：`NineSong/api/docs/client_examples/MediaLibrarySyncMechanism.md`

## Postman

仓库根目录下的公开 Postman 集合文件是：

- `NineSong API.postman_collection.json`
