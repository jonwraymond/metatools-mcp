# PRD-173: Create toolcontent

**Phase:** 7 - Protocol Layer
**Priority:** High
**Effort:** 8 hours
**Dependencies:** PRD-120

---

## Objective

Create `toolprotocol/content/` for unified content/part abstraction across protocols.

---

## Package Contents

- Content type abstraction
- Text, image, resource content
- MIME type handling
- Content streaming

---

## Key Implementation

```go
package content

// Content represents response content.
type Content interface {
    Type() ContentType
    MimeType() string
    Bytes() ([]byte, error)
}

// ContentType defines content types.
type ContentType string

const (
    TypeText     ContentType = "text"
    TypeImage    ContentType = "image"
    TypeResource ContentType = "resource"
    TypeAudio    ContentType = "audio"
    TypeFile     ContentType = "file"
)

// TextContent is text-based content.
type TextContent struct {
    Text string
}

// ImageContent is image-based content.
type ImageContent struct {
    Data     []byte
    MIMEType string
    URI      string
}

// ResourceContent references a resource.
type ResourceContent struct {
    URI      string
    MIMEType string
    Text     string
    Blob     []byte
}

// Builder builds content.
type Builder struct{}

func (b *Builder) Text(text string) Content
func (b *Builder) Image(data []byte, mimeType string) Content
func (b *Builder) Resource(uri string) Content
```

---

## Commit Message

```
feat(content): add content abstraction

Create content package for unified content handling.

Features:
- Content type abstraction
- Text, image, resource types
- MIME type handling
- Content builder

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- PRD-174: Create tooltask
