package main

import (
	"strings"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10

	noFrame = -1
)

func (f *filter) EncodeHeaders(headers api.ResponseHeaderMap, endStream bool) api.StatusType {
	connection, _ := headers.Get("Connection")
	upgrade, _ := headers.Get("Upgrade")
	if strings.ToLower(connection) == "upgrade" && strings.ToLower(upgrade) == "websocket" {
		f.isWebsocket = true
	}
	return api.Continue
}

func (f *filter) DecodeData(data api.BufferInstance, endStream bool) api.StatusType {
	// ignore non-websocket requests
	if !f.isWebsocket {
		return api.Continue
	}

	bytes := data.Bytes()
	f.reqBuffer = append(f.reqBuffer, bytes...)
	var fr *frame
	f.reqBuffer, fr = readFrame(f.reqBuffer)
	if fr == nil {
		// already cache into Golang side
		data.Reset()
		return api.Continue
	}

	// TODO: assume the non-text frame is in a single data buffer
	if fr.frameType != TextMessage {
		return api.Continue
	}

	api.LogDebugf("frame old data: %v", string(fr.GetData()))

	if f.config.action == "remote" {
		go func() {
			bytes := fr.GetData()
			ok := checkData(bytes)
			if !ok {
				bytes = []byte("Unauthorized")
			} else {
				bytes = []byte("Authorized")
			}
			fr.SetData(bytes)
			data.Set(fr.Bytes())
			f.callback.DecoderFilterCallbacks().Continue(api.Continue)
		}()
		return api.Running
	}

	newData := append([]byte("Hello, "), fr.GetData()...)
	fr.SetData(newData)
	data.Set(fr.Bytes())

	api.LogDebugf("frame new data: %v", string(newData))

	return api.Continue
}

func (f *filter) EncodeData(data api.BufferInstance, endStream bool) api.StatusType {
	// ignore non-websocket requests
	if !f.isWebsocket {
		return api.Continue
	}

	bytes := data.Bytes()
	f.rspBuffer = append(f.rspBuffer, bytes...)
	var fr *frame
	f.rspBuffer, fr = readFrame(f.rspBuffer)
	if fr == nil {
		// already cache into Golang side
		data.Reset()
		return api.Continue
	}

	// TODO: assume the non-text frame is in a single data buffer
	if fr.frameType != TextMessage {
		return api.Continue
	}

	newData := append(fr.GetData(), []byte(", World")...)
	fr.SetData(newData)
	data.Set(fr.Bytes())

	return api.Continue
}
