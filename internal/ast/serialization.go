package ast

type SerializationMethod string

const (
	SerializationMethodJSON        SerializationMethod = "json"
	SerializationMethodRAW         SerializationMethod = "raw"
	SerializationMethodMultipart   SerializationMethod = "multipart"
	SerializationMethodForm        SerializationMethod = "form"
	SerializationMethodString      SerializationMethod = "string"
	SerializationMethodEventStream SerializationMethod = "eventstream"
	SerializationMethodJsonL       SerializationMethod = "jsonl"
)
