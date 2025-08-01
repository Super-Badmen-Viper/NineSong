# NineSong Personal Digital Center
NineSong aims to provide cloud native and AI extended solutions for data sharing in various ToB and ToC businesses,
for managing various file metadata (local files, cloud storage files) and the business attributes derived from these metadata, 
and applied to various application scenarios, 
including but not limited to music, movies, notes, documents, photo albums, e-book readers, etc. 
Its goal is to become a representative work in the Github cloud native field.

A Go (Golang) Backend Clean Architecture project with Gin, MongoDB, JWT Authentication Middleware, Test, and Docker.

For NineSong, it only needs to respond to users around the world who need it and always maintain an open source status. You can use any product you like, whether it's following NineSong or choosing a seemingly famous but actually outdated product

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Super-Badmen-Viper/NineSong)

[NSMusicS](https://github.com/Super-Badmen-Viper/NSMusicS):https://github.com/Super-Badmen-Viper/NSMusicS will support Docker deployment by default and integrate the NineSong streaming service.    

## DownLoad For 1.0.0 Version (Music Scene): Released At the end of July 2025
You need to put the. env and docker-compose.yaml files in the same folder. You can customize the parameter configuration of. env and docker-compose.yaml, such as mapping the media library folder to the Volumes of the NineSong container.  
[download](https://github.com/Super-Badmen-Viper/NineSong/releases/): https://github.com/Super-Badmen-Viper/NineSong/releases
Compared to other music servers (such as Navidrome, Jellyfin, Emby, Plex, Subsonic, Gonic), it offers the following enhanced features:
- More comprehensive music library management:
- - [x] Rich single-level sorting options, supporting multi-level mixed sorting and multi-level mixed filtering;
- - [x] deeper processing of composite tags to make the relevance between musics more comprehensive;
- - [x] search jump optimization
- - - [x] Support fuzzy search based on title, album, artist, and lyrics (multiple mixed matching of Chinese Pinyin and simplified traditional Chinese characters);
- - - [x] recommended similar search results;
- More comprehensive music playback experience:
- - [x] various elegant playback styles[cover Square、cover Rotate、cover Beaut、cover Base、cover AlbumList];
- - [x] exclusive playback modes for various music files[normal model、cue-music model];
- CUE exclusive playback (CUE: wav、ape、flac) and CUE file management:
- - [x] Exclusive management page for music disc image (mirror) auxiliary files. 
- - [x] CUE playback styles suitable for music disc image features
- - [x] Visualized virtual track playback of CUE
- More complete TAG import and management:
- - [x] support for importing complete TAGs from more types of music files (including m4a、cue(wav、ape、flac));
- Personalized music recommendations based on user usage data: 
- - [x] Phase 1 (June): Add tag cloud and recommend music based on user interests.
## Subsequent updates (Music Scene):
- More comprehensive music library management:
- - [ ] support for dual-page browsing mode (unlimited virtual list, paged list);
- - [ ] support uploading, downloading, and synchronizing music files between the server and client;
- More complete TAG import and management:
- - [ ] support for user-visualized TAG management, allowing remote uploads, auto-associating, manual merging of artist-album-single TAGs;
- - [ ] support for richer TAG fields: artist profile pictures, artist photos (multiple selection), album covers, song quality versions (multiple selection), and lyrics versions (single selection);
- ISO exclusive playback and ISO file management:
- - [ ] Exclusive management page for music disc image (mirror) auxiliary files.
- - [ ] ISO playback styles suitable for music disc image features
- - [ ] Visualized virtual track playback of ISO
- Support more sound effects settings: 
- - [ ] support for multi-channel audio effects 
- - [ ] support for Advanced/Standard/Simple EQ 
- Integrated free public welfare music TAG API: 
- - [ ] allowing users to obtain online TAGs for songs and choose whether to synchronize TAG data.
- Personalized music recommendations based on user usage data: 
- - [ ] Phase 2 (August): Use lightweight recommendation algorithms based on usage data.
- - [ ] Phase 3 (October): Build a music knowledge graph by analyzing music metadata to achieve smarter recommendations.
- - [ ] Phase 4 (December): Combine the knowledge graph with LLM (DeepSeek) for advanced music recommendations.

## How to Deploy Docker:
You first need to download the compressed file from the [releases](https://github.com/Super-Badmen-Viper/NineSong/releases/)  

You need to put the. env and docker-compose.yaml files [in the same folder](https://github.com/Super-Badmen-Viper/NineSong/releases/). You can customize the parameter configuration of. env and docker-compose.yaml, such as mapping the media library folder to the Volumes of the NineSong container.  

Note that if you update the mirrored version of NineSong, temporary resources in the media library (such as album covers) will also be deleted. You need to rescan the media library in the settings to regenerate temporary resources
```sh
run: docker compose up -d
login mail: admin@gmail.com
login password: admin123
```
How to Thoroughly Reinstall NineSong: Need to Clear Data Together with Volumes in Docker.  
Because considering that the image upgrade cannot affect the database data, if you delete the containers of NineSong, the data in their databases will not disappear unless you clear it together with the data in Volumes in Docker.


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

## Group Chat
- QQ群聊
    - NSMusicS交流群（1）：228440692
- Other | None
- 请注意，所有聊天组仅用于日常沟通、功能请求、需求报告和错误报告，只要您态度端正并支持九歌，在我力所能及的情况下，您的需求我都将尽力满足，但是这需要功能需求排期开发，但是您也可以通过赞助等等开源贡献方式来加快你的需求实现。
- 请不要在聊天组与我进行技术辩论或问答，不然我会给你踢出群聊，技术辩论或问答是非常不礼貌的行为，尤其是在我和你不熟、且你没有参与开源贡献的情况下。
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
