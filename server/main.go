package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gofiber/fiber/v2"
)

const (
	SAMPLE_RATE = 44100
	SECONDS     = 15
)

type (
	Track string

	TrackInfo struct {
		Format string `json:"format"`
		Length int64  `json:"length"`
	}

	Tracks map[Track]TrackInfo

	Params struct {
		File    Track `query:"file"`
		StartAt int64 `query:"start_at"`
	}

	Output struct {
		TrackData   []byte `json:"track_data"`
		TrackLength int64  `json:"track_length"`
		Next        int64  `json:"next"`
	}
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		params := new(Params)

		if err := c.QueryParser(params); err != nil {
			panic(err)
		}

		if params.File == "" {
			return c.JSON(fiber.Map{
				"errors": "file name for download is required",
			})
		}

		conf, err := os.ReadFile("./tracks/tracks.json")
		if err != nil {
			panic(err)
		}

		var tracks Tracks
		if err := json.Unmarshal(conf, &tracks); err != nil {
			panic(err)
		}

		trackInfo, ok := tracks[params.File]
		if !ok {
			return c.Status(404).JSON(fiber.Map{
				"errors": "track dosen't exists",
			})
		}

		decoder, err := mpg123.NewDecoder("")
		chk(err)

		chk(decoder.Open(fmt.Sprintf("./tracks/%s.%s", params.File, trackInfo.Format)))
		defer decoder.Close()

		// get audio format information
		rate, channels, _ := decoder.GetFormat()

		// make sure output format does not change
		decoder.FormatNone()
		decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

		fifhteenSeconds := make([]byte, SECONDS*rate)
		n, err := decoder.Read(fifhteenSeconds)
		if err == mpg123.EOF {
			return c.Status(200).JSON(Output{
				TrackData: make([]byte, 0),
			})
		}
		chk(err)

		c.Set("Content-Type", "audio/mpeg")
		return c.Status(200).JSON(Output{
			TrackData:   fifhteenSeconds[:n],
			TrackLength: trackInfo.Length,
			Next:        params.StartAt + SECONDS,
		})
	})

	app.Listen(":3000")
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
