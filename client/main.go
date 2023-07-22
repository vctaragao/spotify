package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gordonklaus/portaudio"
)

const (
	URL         = "http://localhost:3000/"
	SAMPLE_RATE = 44100
	SECONDS     = 15
)

type trackResponse struct {
	TrackData []byte `json:"track_data"`
}

var mut sync.Mutex

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()

	audio := make([]byte, 2*SAMPLE_RATE)
	sampleFrame := make([]int16, SAMPLE_RATE)

	// go getTrack(audio)
	// go getLocalTrack(audio)

	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)

	chk(decoder.Open("../server/tracks/" + "Daydream - Soobin Hoang SonThaoboy (Hiderway Remix).mp3"))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	go readAsync(decoder, audio)

	stream, err := portaudio.OpenDefaultStream(0, 2, SAMPLE_RATE, len(sampleFrame), &sampleFrame)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	defer stream.Stop()

	for {
		mut.Lock()
		fmt.Println("1. Reading from buffer")
		fmt.Println(audio)
		err := binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, sampleFrame)
		fmt.Println("SAMPLE FRAME:")
		fmt.Println(sampleFrame)
		mut.Unlock()

		chk(err)
		chk(stream.Write())

		select {
		case <-sig:
			return
		default:
		}
	}
}

func getTrack(buffer []byte) {
	file := "Daydream%20-%20Soobin%20Hoang%20SonThaoboy%20%28Hiderway%20Remix%29"

	resp, err := http.Get(fmt.Sprintf("%s?file=%s", URL, file))
	chk(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	chk(err)

	var track trackResponse
	json.Unmarshal(data, &track)

	fmt.Println("writting track to buffer")

	fmt.Println(len(track.TrackData))

	chk(binary.Write(bytes.NewBuffer(track.TrackData), binary.LittleEndian, buffer))
}

func getLocalTrack(buffer []byte) {
	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)

	chk(decoder.Open("../server/tracks/" + "Daydream - Soobin Hoang SonThaoboy (Hiderway Remix).mp3"))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)
	fmt.Println("writting track to buffer")
	for {
		n, err := decoder.Read(buffer)
		if err == mpg123.EOF {
			return
		}
		chk(err)
		fmt.Printf("bytes wriiten: %v\n", n)
	}
}

func readAsync(decoder *mpg123.Decoder, buffer []byte) {
	for {
		mut.Lock()

		fmt.Println("writting track to buffer")
		_, err := decoder.Read(buffer)
		if err == mpg123.EOF {
			break
		}
		chk(err)

		mut.Unlock()
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
