// Package tag provides primitives for generating tags for transaction
package tag

import (
	"encoding/binary"
	"errors"

	"github.com/linkedin/goavro/v2"
	"github.com/liteseed/goar/crypto"
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

// fromAvro encoded binary data convert to human-readable [Tag]
func fromAvro(data []byte) (*[]Tag, error) {
	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}

	avroTags, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return nil, err
	}

	var tags []Tag

	for _, v := range avroTags.([]any) {
		tag := v.(map[string]any)
		tags = append(tags, Tag{Name: string(tag["name"].([]byte)), Value: string(tag["value"].([]byte))})
	}
	return &tags, err
}

// toAvro convert from human-readable [Tag] to Avro encoded binary data
func toAvro(tags *[]Tag) ([]byte, error) {
	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}

	var avroTags []map[string]any

	for _, tag := range *tags {
		m := map[string]any{"name": []byte(tag.Name), "value": []byte(tag.Value)}
		avroTags = append(avroTags, m)
	}
	data, err := codec.BinaryFromNative(nil, avroTags)
	if err != nil {
		return nil, err
	}
	return data, err
}

// Serialize Converts readable Tag data into avro-encoded byte data for an Arweave transaction
// Learn more: https://github.com/ArweaveTeam/arweave-standards/blob/master/ans/ANS-104.md
func Serialize(tags *[]Tag) ([]byte, error) {
	if len(*tags) > 0 {
		data, err := toAvro(tags)
		if err != nil {
			return nil, err
		}

		return data, nil
	}
	return nil, nil
}

// Deserialize Convert avro-encoded byte data from an Arweave transaction into readable Tag data
// Learn more: https://github.com/ArweaveTeam/arweave-standards/blob/master/ans/ANS-104.md
func Deserialize(data []byte, startAt int) (*[]Tag, int, error) {
	tags := &[]Tag{}
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

		tags, err := fromAvro(bytesData)
		if err != nil {
			return nil, tagsEnd, err
		}
		tagsEnd = bytesDataEnd
		return tags, tagsEnd, nil
	}
	return tags, tagsEnd, nil
}

// Decode convert [Tag] to bytes
func Decode(tags *[]Tag) ([][][]byte, error) {
	if len(*tags) == 0 {
		return nil, nil
	}
	data := make([][][]byte, 0)
	for _, tag := range *tags {
		name, err := crypto.Base64URLDecode(tag.Name)
		if err != nil {
			return nil, err
		}
		value, err := crypto.Base64URLDecode(tag.Value)
		if err != nil {
			return nil, err
		}
		data = append(data, [][]byte{name, value})
	}
	return data, nil
}

// ConvertToBase64 encode all string values of [Tag] to Base64 string
func ConvertToBase64(tags *[]Tag) *[]Tag {
	var result []Tag
	for _, tag := range *tags {
		result = append(result, Tag{Name: crypto.Base64URLEncode([]byte(tag.Name)), Value: crypto.Base64URLEncode([]byte(tag.Value))})
	}
	return &result
}
