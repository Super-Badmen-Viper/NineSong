# NineSong Server
NineSong aims to provide cloud native and AI extended solutions for data sharing in various ToB and ToC businesses,
for managing various file metadata (local files, cloud storage files) and the business attributes derived from these metadata, 
and applied to various application scenarios, 
including but not limited to music, movies, notes, documents, photo albums, e-book readers, etc. 
Its goal is to become a representative work in the Github cloud native field.

A Go (Golang) Backend Clean Architecture project with Gin, MongoDB, JWT Authentication Middleware, Test, and Docker.

For NineSong, it only needs to respond to users around the world who need it and always maintain an open source status. You can use any product you like, whether it's following NineSong or choosing a seemingly famous but actually outdated product

## New Function
NineSong Server will be released in June this year.  
[NSMusicS](https://github.com/Super-Badmen-Viper/NSMusicS):https://github.com/Super-Badmen-Viper/NSMusicS will support Docker deployment by default and integrate the NineSong streaming service.  
Compared to other music servers (such as Navidrome, Jellyfin, Emby, Plex, Subsonic, Gonic), it offers the following enhanced features:
- More comprehensive music library management:
- - Rich single-level sorting options, supporting multi-level mixed sorting and multi-level mixed filtering;
- - deeper processing of composite tags to make the relevance between musics more comprehensive;
- - search jump optimization
- - - support for Chinese pinyin fuzzy search;
- - - support searching based on lyrics;
- - - support for quick initial letter jumping;
- - - recommended similar search results;
- - support for dual-page browsing mode (unlimited virtual list, paged list);
- - support uploading, downloading, and synchronizing music files between the server and client;
- More comprehensive music playback experience:
- - various elegant playback styles[cover Square、cover Rotate、cover Beaut、cover Base、cover AlbumList];
- - exclusive playback modes for various music files[normal model、cue-music model];
- CUE exclusive playback (CUE: wav、ape、flac) and CUE file management:
- - Exclusive management page for music disc image (mirror) auxiliary files. 
- - CUE playback styles suitable for music disc image features
- - Visualized virtual track playback of CUE
- More complete TAG import and management:
- - support for importing complete TAGs from more types of music files (including m4a、cue(wav、ape、flac));
- - support for user-visualized TAG management, allowing remote uploads, auto-associating, manual merging of artist-album-single TAGs;
- - support for richer TAG fields: artist profile pictures, artist photos (multiple selection), album covers, song quality versions (multiple selection), and lyrics versions (single selection);
- Integrated free public welfare music TAG API, allowing users to obtain online TAGs for songs and choose whether to synchronize TAG data.
- Support for multi-channel audio effects; support for Advanced/Standard/Simple EQ; (October)
- Personalized music recommendations based on user usage data.
- - Phase 1 (June): Add tag cloud and recommend music based on user interests.
- - Phase 2 (August): Use lightweight recommendation algorithms based on usage data.
- - Phase 3 (October): Build a music knowledge graph by analyzing music metadata to achieve smarter recommendations.
- - Phase 4 (December): Combine the knowledge graph with LLM (DeepSeek) for advanced music recommendations.

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

The Chinese name of the project is "Nine Song | 九歌", abbreviated as NSMusicS<br> inspired by ["Chu Ci"] | 楚辞, to commemorate ["Qu Yuan"] | 屈原<br>
