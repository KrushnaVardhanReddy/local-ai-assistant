package stt

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type GroqEngine struct {
	apiKey string
	model  string
}

// NewGroqEngine creates a new STTEngine that uses the Groq Cloud API
func NewGroqEngine(apiKey string, model string) *GroqEngine {
	if model == "" {
		model = "whisper-large-v3-turbo" // Default fast model
	}
	return &GroqEngine{
		apiKey: apiKey,
		model:  model,
	}
}

func (g *GroqEngine) TranscribeStream(audio []float32) (chan string, error) {
	if len(audio) == 0 {
		return nil, errors.New("empty audio")
	}

	ch := make(chan string)

	go func() {
		defer close(ch)

		wavBytes := createWAV(audio, 16000)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", "audio.wav")
		if err == nil {
			part.Write(wavBytes)
		}
		writer.WriteField("model", g.model)
		writer.WriteField("response_format", "json")
		writer.Close()

		req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
		if err != nil {
			fmt.Printf("[Groq STT] Request error: %v\n", err)
			return
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+g.apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[Groq STT] Network error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[Groq STT] API error (%d): %s\n", resp.StatusCode, string(respBody))
			return
		}

		var result struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(respBody, &result); err == nil && result.Text != "" {
			ch <- result.Text
		}
	}()

	return ch, nil
}

func (g *GroqEngine) Close() error {
	return nil
}

func createWAV(audio []float32, sampleRate int) []byte {
	buf := new(bytes.Buffer)

	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+len(audio)*2))
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(buf, binary.LittleEndian, uint16(2))
	binary.Write(buf, binary.LittleEndian, uint16(16))

	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(len(audio)*2))

	for _, sample := range audio {
		if sample > 1.0 {
			sample = 1.0
		}
		if sample < -1.0 {
			sample = -1.0
		}
		s16 := int16(sample * 32767.0)
		binary.Write(buf, binary.LittleEndian, s16)
	}

	return buf.Bytes()
}
