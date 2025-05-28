# NineSong Server
NineSong aims to provide cloud native and AI extended solutions for data sharing in various ToB and ToC businesses, used to manage various file metadata and metadata derived business attributes, and applied to various application scenarios, including but not limited to music, movies, notes, documents, photo albums, e-book readers, etc. Its goal is to become a representative work in the Github cloud native field.

A Go (Golang) Backend Clean Architecture project with Gin, MongoDB, JWT Authentication Middleware, Test, and Docker.

## New Function
NineSong Server will be released in June this year.  
[NSMusicS](https://github.com/Super-Badmen-Viper/NSMusicS):https://github.com/Super-Badmen-Viper/NSMusicS will support Docker deployment by default and integrate the NineSong streaming service.  
Compared to other music servers (such as Navidrome, Jellyfin, Emby, Plex, Subsonic, Gonic), it offers the following enhanced features:
- More detailed music library management, including remote upload, synchronization, and download.
- More detailed music metadata TAG management, with support for remote TAG upload and synchronized editing.
- Add TAG settings (artist avatar, artist photos (multiple selections), album cover, song quality version (multiple selections), and lyrics version (single selection)).
- Adaptation of a free music TAG API, allowing users to obtain online TAGs for their songs and choose whether to synchronize TAG data.
- - https://musicbrainz.org/、https://www.theaudiodb.com/
- More detailed music recommendations based on your playback data.
- - Phase 1(June): Add tag clouds to recommend music based on user interests.
- - Phase 2(August): Use lightweight recommendation algorithms based on usage data.
- - Phase 3(October): Build a music knowledge graph for smarter recommendations by analyzing music metadata.
- - Phase 4(December): Combine knowledge graphs with LLM(DeepSeek) for advanced music recommendations.
- A richer music playback experience with a variety of playback modes for a more comprehensive and refined effect.
- Multi-track sound effect settings and transmission(October)

## NineSong | NineSong Multimedia(Server) : 九歌多媒体
- [ ] Compatible with streaming media servers (Jellyfin、Emby、Navidrome、Plex)
- [ ] General file library management(Audio、Video、Image、Text、Document、Archive、Executable、Database、Unknown)
- [ ] Scene of Streaming Music and Karaoke
- [ ] Scene of AI-Models deploy
- [ ] Scene of Intelligent Gallery album
- [ ] Scene of Film and Television Center
- [ ] Scene of Online Notes
- [ ] Scene of Document Workbench
- [ ] Scene of E-book reader
- [ ] Knowledge graph Recommendation system
- [ ] Internationalization

```sh
HTTP请求 → Controller → Usecase（业务逻辑） → Repository（数据操作） → MongoDB  
↑               ↑  
定义接口         实现接口
(Domain)       (Repository)
```

## local debug run
 - modify: .env
   - DB_HOST=localhost
   - DB_PORT=27017
   - if $not local MongoDB_DATA_VOLUME
     - modify: docker-compose-local-windows.ps1 
       - $env:MongoDB_DATA_VOLUME = "C:\Users\Public\Documents\NineSong\MongoDB"
       - $env:SQLITE_DATA_VOLUME = "C:\Users\Public\Documents\NineSong\Sqlite"
       - $env:MUSIC_DATA_VOLUME = "E:\0_Music"
     - run: docker-compose-local-windows.ps1
     - await create: $env:MongoDB_DATA_VOLUME...
   - else if $have local MongoDB_DATA_VOLUME
     - run: docker-compose-mongodb.yaml
 - go install github.com/air-verse/air@latest
 - run: air
   
## docker build run
 - modify: .env
   - DB_HOST=mongodb
   - DB_USER=jiuge01
   - DB_PASS=jiuge01
 - modify: docker-compose.yaml
   - web: volumes
   - mongodb: volumes
 - run: docker-compose.yaml

## postman run
Import postman.json file: NineSong API.postman_collection.json

## Thanks:
 - [go-backend-clean-architecture](https://github.com/amitshekhariitbhu/go-backend-clean-architecture) : NineSong's Clean Architecture Project Template
 - [go-audio](https://github.com/go-audio) : A comprehensive GO audio library with a wide range of data types
 - [go-taglib](https://github.com/sentriz/go-taglib) : GO TAG library with comprehensive coverage of media file types
 - ......