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

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gordonklaus/portaudio"
)

const (
	URL         = "http://localhost:3000/"
	SAMPLE_RATE = 44100
	SECONDS     = 15
)

var sig chan os.Signal = make(chan os.Signal, 1)

type trackResponse struct {
	Next        int64  `json:"next"`
	TrackData   []byte `json:"track_data"`
	TrackLength int64  `json:"track_length"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  input file name")
		return
	}
	fmt.Println("Playing.  Press Ctrl-C to stop.")
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()

	out := make([]int16, SAMPLE_RATE)
	stream, err := portaudio.OpenDefaultStream(0, 2, SAMPLE_RATE, len(out), &out)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	defer stream.Stop()

	chAudio := make(chan []byte, 30*SAMPLE_RATE)

	fmt.Printf("bytes preenchidos na criação do canal de audio: %v\n", len(chAudio))

	// go readAsync(chAudio)
	go getTrack(chAudio)
	for {
		select {
		case audio := <-chAudio:
			fmt.Println("lendo canal de audio")
			fmt.Printf("bytes escritos no chAudio: %v\n", len(audio))
			chk(binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, out))
			fmt.Println("escrevendo no stream de audio")
			chk(stream.Write())
		case <-sig:
			fmt.Println("bye bye")
			return
		default:
			fmt.Println("default")
		}
	}
}

func readAsync(chAudio chan []byte) {
	decoder, err := mpg123.NewDecoder("")
	chk(err)

	fileName := os.Args[1]
	chk(decoder.Open("../server/tracks/" + fileName))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	for {
		audio := make([]byte, 2*SAMPLE_RATE)
		_, err := decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		chk(err)
		chAudio <- audio
	}
}

// func readAsyncFromTrack(chTrack, chAudio chan []byte) {
// decoder, err := mpg123.NewDecoder("")
// chk(err)

// fileName := os.Args[1]
// chk(decoder.Open("../server/tracks/" + fileName))
// defer decoder.Close()

// // get audio format information
// rate, channels, _ := decoder.GetFormat()

// // make sure output format does not change
// decoder.FormatNone()
// decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

// for {
// select {
// case track := <-chTrack:
// audio := make([]byte, 2*SAMPLE_RATE)
// _, err := decoder.Read(audio)
// if err == mpg123.EOF {
// break
// }
// chk(err)
// chAudio <- audio
// case <-sig:
// fmt.Println("bye bye")
// return
// default:
// }
// }
// }

func getTrack(chAudio chan []byte) {
	startAt := 0
	file := "Daydream%20-%20Soobin%20Hoang%20SonThaoboy%20%28Hiderway%20Remix%29"

	for {
        url := fmt.Sprintf("%s?file=%s&start_at=%s", URL, file, fmt.Sprint(startAt))
		resp, err := http.Get(url)
		chk(err)
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		chk(err)

		var track trackResponse
		json.Unmarshal(data, &track)

        fmt.Println("writting track to buffer")
        fmt.Println(len(track.TrackData))

		if len(track.TrackData) == 0 || track.Next >= track.TrackLength {
			break
		}

		chAudio <- track.TrackData
		startAt = int(track.Next)
	}

	// chk(binary.Write(bytes.NewBuffer(track.TrackData), binary.LittleEndian, chAudio))
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
