# NineSong Server
NineSong Goal is to achieve various application scenarios (such as music, videos, movies, notes, photo albums, documents, books, etc.), and provide internationalization, cloud native deployment, streaming services, fine metadata management, and cross platform data management for various application scenarios, becoming a representative work in the Github cloud native field

A Go (Golang) Backend Clean Architecture project with Gin, MongoDB, JWT Authentication Middleware, Test, and Docker.

## New Function
The NineSong Server will be released in June this year, and NSMusicS will be able to deploy Docker and integrate NineSong streaming services by default.  
Compared to other music servers (Navidrome, Jellyfin, Emby, Plex, Subsonic, Gonic),
It has better functions, as follows:
- More detailed music library managementand support for remote upload, synchronization, and download of music libraries
- More detailed music metadata TAG management, supporting remote TAG upload and synchronous modification
- - Add TAG settings (singer avatar, singer photo (multiple choices), album cover, song sound quality version (multiple choices), song lyrics version (multiple choices))
- - Added adaptation to the free music TAG API, allowing users to obtain the online TAG of their songs and choose whether to synchronize TAG data。https://musicbrainz.org/、https://www.theaudiodb.com/
- More detailed music recommendations tailored to your playback data
- - Add tag word cloud for users to recommend interests
- More detailed music playback experience, with a variety of playback modes available
- It can be used through a web browser to automatically respond to changes in the layout of computer and mobile applications based on screen ratio

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