package docsData

import (
	"encoding/json"
)

type TagData struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description"`
	OperationChunkIds []string `json:"operationChunkIds"`
}

type TagChunk struct {
	ID        string  `json:"id"`
	Slug      string  `json:"slug"`
	ChunkData TagData `json:"chunkData"`
	ChunkType string  `json:"chunkType"`
}

func SerializeTag(docs *Docs, tag string, description *string, operationIds []string) error {
	if tag == "" {
		tag = "Global"
	}
	chunk := TagChunk{
		ID:   docs.GenerateID(),
		Slug: "endpoint/" + tag,
		ChunkData: TagData{
			Name:              tag,
			Description:       description,
			OperationChunkIds: operationIds,
		},
		ChunkType: "tag",
	}
	b, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	docs.SaveSerializedChunk(chunk.ID, string(b))
	return nil
}
