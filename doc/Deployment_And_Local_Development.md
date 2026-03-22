# Deployment And Local Development

This document preserves the operational notes that used to live in the old public README, while aligning the file names with the current repository layout.

## Public Docker Deployment

Public release packages were described as being available from:

- https://github.com/Super-Badmen-Viper/NineSong/releases/

The archived deployment flow was:

1. Download the release package.
2. Keep `.env` and `docker-compose.yaml` in the same folder.
3. Adjust environment values and volume mappings as needed, especially the host path for your media library.
4. Run:

```sh
docker compose up -d
```

## Default Bootstrap Credentials In This Snapshot

The code in `NineSong/bootstrap/init.go` contains the default bootstrap account shown below:

- Login email: `admin@gmail.com`
- Login password: `admin123`

## Important Historical Notes

- Updating the mirrored NineSong version could remove temporary generated resources in the media library, such as album covers.
- A media-library rescan could be required to regenerate those temporary resources.
- A full reinstall required clearing both containers and Docker volume data together.

## Local Development References

The current repository layout exposes these useful files for local work:

- Windows local init script: `NineSong/docker-compose-local-develop-init-windows.ps1`
- Local MongoDB compose file: `NineSong/docker-compose-local-mongo.yaml`
- Main compose file: `NineSong/docker-compose.yaml`

The old README suggested these local environment adjustments:

- `DB_HOST=localhost`
- `DB_PORT=27017`

The Windows init script in this repository currently prepares these default volume paths:

- `C:\Users\Public\Documents\NineSong\MongoDB`
- `C:\Users\Public\Documents\NineSong\Sqlite`
- `C:\Users\Public\Documents\NineSong\MetaData`
- `E:\0_Music`

For live local development, the archived README also pointed to:

```sh
go install github.com/air-verse/air@latest
air
```

## API And Client Documentation Already Included In The Repo

- Index: `NineSong/api/docs/README.md`
- Full API docs: `NineSong/api/docs/API_Documentation.md`
- Feature summary: `NineSong/api/docs/FeatureSummary.md`
- Quick start: `NineSong/api/docs/client_examples/QuickStartGuide.md`
- Client implementation guide: `NineSong/api/docs/client_examples/MediaLibraryClientImplementationGuide.md`
- API examples: `NineSong/api/docs/client_examples/MediaLibraryAPIUsageExamples.md`
- Sync mechanism guide: `NineSong/api/docs/client_examples/MediaLibrarySyncMechanism.md`

## Postman

The public Postman collection at the repository root is:

- `NineSong API.postman_collection.json`
