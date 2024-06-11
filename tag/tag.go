package tag

import (
	"encoding/binary"
	"errors"

	"github.com/linkedin/goavro/v2"
)

const avroTagSchema = `
{
	"type": "array",
	"items": {
		"type": "record",
		"name": "Tag",
		"fields": [
			{ "name": "name", "type": "bytes" },
			{ "name": "value", "type": "bytes" }
		]
	}
}`

type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func decodeAvro(data []byte) ([]Tag, error) {
	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}

	avroTags, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return nil, err
	}

	tags := []Tag{}

	for _, v := range avroTags.([]any) {
		tag := v.(map[string]any)
		tags = append(tags, Tag{Name: string(tag["name"].([]byte)), Value: string(tag["value"].([]byte))})
	}
	return tags, err
}

func encodeAvro(tags []Tag) ([]byte, error) {
	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}

	avroTags := []map[string]any{}

	for _, tag := range tags {
		m := map[string]any{"name": []byte(tag.Name), "value": []byte(tag.Value)}
		avroTags = append(avroTags, m)
	}
	data, err := codec.BinaryFromNative(nil, avroTags)
	if err != nil {
		return nil, err
	}
	return data, err
}

func Decode(tags []Tag) ([]byte, error) {
	if len(tags) > 0 {
		data, err := encodeAvro(tags)
		if err != nil {
			return nil, err
		}

		return data, nil
	}
	return nil, nil
}

func Encode(data []byte, startAt int) ([]Tag, int, error) {
	tags := []Tag{}
	tagsEnd := startAt + 8 + 8
	numberOfTags := int(binary.LittleEndian.Uint16(data[startAt : startAt+8]))
	numberOfTagBytesStart := startAt + 8
	numberOfTagBytesEnd := numberOfTagBytesStart + 8
	numberOfTagBytes := int(binary.LittleEndian.Uint16(data[numberOfTagBytesStart:numberOfTagBytesEnd]))
	if numberOfTags > 127 {
		return tags, tagsEnd, errors.New("invalid data item - max tags 127")
	}
	if numberOfTags > 0 && numberOfTagBytes > 0 {
		bytesDataStart := numberOfTagBytesEnd
		bytesDataEnd := numberOfTagBytesEnd + numberOfTagBytes
		bytesData := data[bytesDataStart:bytesDataEnd]

		tags, err := decodeAvro(bytesData)
		if err != nil {
			return nil, tagsEnd, err
		}
		tagsEnd = bytesDataEnd
		return tags, tagsEnd, nil
	}
	return tags, tagsEnd, nil
}
