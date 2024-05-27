package network

import (
	"encoding/json"
	"log"
)

type Message interface {
	Dest() int // Dest returns the destination of the message
	From() int // From returns the source of the messsge
	Size() int // Size returns the size of the object in the emulator.
}

type HttpMessage struct {
	DataType string
	Content  []byte
}

type JsonMessage struct {
	DataType string `json:"dataType"`
	Content  []byte `json:"content"`
}

type ContentData struct {
	Dest int    `json:"dest"`
	From int    `json:"from"`
	Size int    `json:"size"`
	Data string `json:"data"`
}

func (jm JsonMessage) Dest() int {
	var contentData ContentData

	if err := json.Unmarshal(jm.Content, &contentData); err != nil {
		log.Printf("Failed to unmarshal content: %v", err)
		return -1 // -1 : nil
	}
	return contentData.Dest
}

func (jm JsonMessage) From() int {
	var contentData ContentData
	if err := json.Unmarshal(jm.Content, &contentData); err != nil {
		log.Printf("Failed to unmarshal content: %v", err)
		return -1 // -1 : nil
	}
	return contentData.From
}

func (jm JsonMessage) Size() int {
	var contentData ContentData
	if err := json.Unmarshal(jm.Content, &contentData); err != nil {
		log.Printf("Failed to unmarshal content: %v", err)
		return -1
	}
	return contentData.Size
}

func (jm *JsonMessage) SetContent(dest, from int, data string) {
	contentData := ContentData{
		Dest: dest,
		From: from,
		Data: data,
	}
	contentBytes, err := json.Marshal(contentData)
	if err != nil {
		log.Printf("Failed to marshal content data: %v", err)
		return
	}
	jm.Content = contentBytes
}
