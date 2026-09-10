package docsData

import "fmt"

// This file defines some base operations we use across all chunks

type Docs struct {
	// Maps from slug to ID
	slugMap   map[string]string
	chunkList []*string
	id        int
}

func (d *Docs) GenerateID() string {
	d.id++
	return fmt.Sprintf("$%d", d.id)
}

// TODO: This is something of a quick-n-dirty way to record and check if we've
// processed an chunk before, to prevent infinite recursion. It's not great
// though since not all chunks have slugs. We'll need to fix a few issues to
// make this more robust though, so I'm punting for the time being.
func (d *Docs) RegisterSlug(id string, slug string) {
	d.slugMap[slug] = id
}
func (d *Docs) GetCachedIdFromSlug(slug string) string {
	id, ok := d.slugMap[slug]
	if !ok {
		return ""
	}
	return id
}

// Save a serialized chunk so that we can later save it to disk. One day, we can
// also use this to immediately send a chunk to the renderer over IPC/etc.
func (d *Docs) SaveSerializedChunk(id string, serializedChunk string) {
	d.chunkList = append(d.chunkList, &serializedChunk)
}
