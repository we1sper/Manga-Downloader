# ManHuaGui Server APIs

## Query Manga

Method: `GET`

Path: `/query/manga`

Query Parameters:

- `mid`: the manga id, unsigned int

Response:

```json
{
    "id": 0,            // the manga id, unsigned int
    "name": "",         // the manga name
    "author": "",       // the manga author
    "introduction": "", // the introduction of the manga
    "date": "",         // the published date
    "status": "",       // the status, e.g., "连载中"或"已完结"
    "cover": "",        // the url of cover image
    "contents": [       // the chapter list
        {
            "title": "", // the chapter title
            "href": "",  // the link to the chapter
            "page": ""   // the page info of the chapter, e.g., "128p"
        }
    ]
}
```

## Download Chapters

Method: `POST`

Path: `/download/chapters`

Request Body:

```json
[
    {
        "mid": 0, // the manga id, unsigned int
        "cid": 0  // the chapter id, unsigned int
    }
]
```

Response:

```json
[
    {
        "mid": 0,    // the manga id, unsigned int
        "cid": 0,    // the chapter id, unsigned int
        "status": "" // task status, e.g., "error: <cause>" or "ok"
    }
]
```

## Query Download Records

Method: `GET`

Path: `/query/records`

Response:

```json
[
    {
        "mid": 0,        // the manga id
        "mname": "",     // the manga name
        "cid": 0,        // the chapter id
        "cname": "",     // the chapter name
        "progress": 0.0, // the task progress, ranges from 0.0 to 100.0
        "status": "",    // the task status, e.g., "waiting for download", "downloading", "success" and "error: <cause>"
        "count": 0,      // the count of completed downloaded files of the chapter 
        "total": 0       // the number of all files of the chapter
    }
]
```
