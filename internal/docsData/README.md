# Docs Data

The `docsData` package is responsible for creating a data structure that can be used to render docs. It does this by serializing the AST into a series of chunks, which are then sent to an external rendering process for conversion to docs pages. Docs are typically sent to the rendering process via JSONL files, but could also be sent over IPC or over a network connection.

## Chunks

Core to this package is the concept of a “chunk.” A chunk is a piece of data that can be rendered into a piece of UI. Chunks are fairly similar to refs in OpenAPI, except that _everything_ is represented as a chunk. Like refs, chunks can reference other chunks.

Some refs will correlate 1:1 with chunks, and vice versa, but this is not uniformly the case. Refs are optimized to help authors write specs more efficiently, i.e. they’re optimized for humans. Chunks however are optimized for later stages of rendering, i.e. they’re optimized for machines. As such, there are different needs of each.

Every chunk has the following base data structure:

```json
{
  "id": "string",
  "slug": "string",
  "chunkType": "string",
  "chunkData": "object"
}
```

- `id` - a unique id that represents this chunk. There is no semantic meaning to ids, other than that they are unique
- `slug` - if non-empty, a unique URL friendly slug that represents this chunk. It may or may not be used as a URL in generated docs.
- `chunkType` - a string identifying the chunk type. We use this during deserialization to know how to interpret the chunk
- `chunkData` - the chunk specific data

Some examples of chunks:
- `about` - a chunk that contains high-level information about the entire spec, and mostly comes from the `info` section of the spec
- `schema` - a chunk that contains information about a schema, and is used to represent any piece of data used by operations, web hooks, etc.

Type definitions for each chunk are declared in the implementation file that processes each chunk.
