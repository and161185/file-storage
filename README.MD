# File-Storage

Production-oriented file storage service written in Go for use in a corporate network.

Primary use cases:
- product images for marketplaces
- internal documents and media files

The service supports controlled public access for selected files.

---

## Features

- Store and retrieve files by ID
- Idempotent delete; upsert is idempotent when an explicit file ID is provided
- Image processing (resize)
- Public / private access model
- Filesystem-based storage (no DB)

---

## Image processing

The service determines whether a file is an image using:
- explicit `is_image` flag in the request, or
- automatic detection from binary data if the flag is omitted

If the file is an image:
- its dimensions are reduced so that the longest side does not exceed  
  `image.max_dimension` from the configuration
- the processed image is stored in the format defined by `image.ext`

Non-image files are stored **as-is**, without modification.

---

## Supported formats

### Input formats
- JPEG
- PNG
- BMP
- GIF
- TIFF
- WebP

### Storage formats for images
- JPEG
- PNG
- BMP
- GIF

---

## Access model

Access control is based on the `app.security` configuration section.

- **Public files**
  - content does not require authorization
- **Non-public files**
  - content requires authorization
- Metadata (`info`) and all write operations always require authorization

Public visibility is a business-level property of a file.

---

## How to build and run

```bash
go build -o bin/filestorage ./cmd
./bin/filestorage --config configs/config.yaml
```

To build with a specific version:
```bash
go build -ldflags "-X main.version=0.1.0" -o bin/filestorage ./cmd
```