package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

type (
	Track string

	TrackInfo struct {
		Format string `json:"format"`
		Length int    `json:"length"`
	}

	Tracks map[Track]TrackInfo

	Params struct {
		File    Track `query:"file"`
		StartAt int   `query:"start_at"`
	}

	Output struct {
		TrackData []byte `json:"track_data"`
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

		f, err := os.Open(fmt.Sprintf("./tracks/%s.%s", params.File, trackInfo.Format))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		bytePerSecond := getTrackBytePerSecond(f, trackInfo)

		fmt.Println(params.StartAt)
		if _, err = f.Seek(bytePerSecond*int64(params.StartAt), 0); err != nil {
			panic(err)
		}

		fifhteenSeconds := make([]byte, 15*bytePerSecond)
		b, err := f.Read(fifhteenSeconds)
		if err != nil {
			panic(err)
		}

		return c.Status(200).JSON(Output{
			TrackData: fifhteenSeconds[:b],
		})
	})

	app.Listen(":3000")
}

func getTrackBytePerSecond(f *os.File, trackInfo TrackInfo) int64 {
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}

	return stat.Size() / int64(trackInfo.Length)
}
