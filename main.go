package main

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdk "github.com/SkYNewZ/streamdeck-sdk"
)

var (
	PlexAmpAddress = "http://localhost:63460"
	PlexAddress    = "http://192.168.1.100:32401"

	defaultThumbnailData []byte

	contexts   = make(map[string]struct{})
	streamdeck *sdk.StreamDeck
)

type Track struct {
	Text              string `xml:",chardata"`
	State             string `xml:"state,attr"`
	Duration          string `xml:"duration,attr"`
	Time              string `xml:"time,attr"`
	PlayQueueItemID   string `xml:"playQueueItemID,attr"`
	Key               string `xml:"key,attr"`
	RatingKey         string `xml:"ratingKey,attr"`
	PlayQueueID       string `xml:"playQueueID,attr"`
	PlayQueueVersion  string `xml:"playQueueVersion,attr"`
	ContainerKey      string `xml:"containerKey,attr"`
	Type              string `xml:"type,attr"`
	ItemType          string `xml:"itemType,attr"`
	Volume            string `xml:"volume,attr"`
	Shuffle           string `xml:"shuffle,attr"`
	Repeat            string `xml:"repeat,attr"`
	Controllable      string `xml:"controllable,attr"`
	MachineIdentifier string `xml:"machineIdentifier,attr"`
	Protocol          string `xml:"protocol,attr"`
	Address           string `xml:"address,attr"`
	Port              string `xml:"port,attr"`
	Track             struct {
		Text                 string `xml:",chardata"`
		PlayQueueItemID      string `xml:"playQueueItemID,attr"`
		RatingKey            string `xml:"ratingKey,attr"`
		Key                  string `xml:"key,attr"`
		ParentRatingKey      string `xml:"parentRatingKey,attr"`
		GrandparentRatingKey string `xml:"grandparentRatingKey,attr"`
		Guid                 string `xml:"guid,attr"`
		ParentGuid           string `xml:"parentGuid,attr"`
		GrandparentGuid      string `xml:"grandparentGuid,attr"`
		ParentStudio         string `xml:"parentStudio,attr"`
		Type                 string `xml:"type,attr"`
		Title                string `xml:"title,attr"`
		GrandparentKey       string `xml:"grandparentKey,attr"`
		ParentKey            string `xml:"parentKey,attr"`
		LibrarySectionTitle  string `xml:"librarySectionTitle,attr"`
		LibrarySectionID     string `xml:"librarySectionID,attr"`
		LibrarySectionKey    string `xml:"librarySectionKey,attr"`
		GrandparentTitle     string `xml:"grandparentTitle,attr"`
		ParentTitle          string `xml:"parentTitle,attr"`
		Summary              string `xml:"summary,attr"`
		Index                string `xml:"index,attr"`
		ParentIndex          string `xml:"parentIndex,attr"`
		RatingCount          string `xml:"ratingCount,attr"`
		ViewCount            string `xml:"viewCount,attr"`
		LastViewedAt         string `xml:"lastViewedAt,attr"`
		ParentYear           string `xml:"parentYear,attr"`
		Thumb                string `xml:"thumb,attr"`
		ParentThumb          string `xml:"parentThumb,attr"`
		GrandparentThumb     string `xml:"grandparentThumb,attr"`
		Duration             string `xml:"duration,attr"`
		AddedAt              string `xml:"addedAt,attr"`
		UpdatedAt            string `xml:"updatedAt,attr"`
		MusicAnalysisVersion string `xml:"musicAnalysisVersion,attr"`
		Source               string `xml:"source,attr"`
		Media                struct {
			Text          string `xml:",chardata"`
			ID            string `xml:"id,attr"`
			Duration      string `xml:"duration,attr"`
			Bitrate       string `xml:"bitrate,attr"`
			AudioChannels string `xml:"audioChannels,attr"`
			AudioCodec    string `xml:"audioCodec,attr"`
			Container     string `xml:"container,attr"`
			Part          struct {
				Text      string `xml:",chardata"`
				ID        string `xml:"id,attr"`
				Key       string `xml:"key,attr"`
				Duration  string `xml:"duration,attr"`
				File      string `xml:"file,attr"`
				Size      string `xml:"size,attr"`
				Container string `xml:"container,attr"`
				Stream    struct {
					Text                 string `xml:",chardata"`
					ID                   string `xml:"id,attr"`
					StreamType           string `xml:"streamType,attr"`
					Selected             string `xml:"selected,attr"`
					Codec                string `xml:"codec,attr"`
					Index                string `xml:"index,attr"`
					Channels             string `xml:"channels,attr"`
					Bitrate              string `xml:"bitrate,attr"`
					AlbumGain            string `xml:"albumGain,attr"`
					AlbumPeak            string `xml:"albumPeak,attr"`
					AlbumRange           string `xml:"albumRange,attr"`
					AudioChannelLayout   string `xml:"audioChannelLayout,attr"`
					BitDepth             string `xml:"bitDepth,attr"`
					EndRamp              string `xml:"endRamp,attr"`
					Gain                 string `xml:"gain,attr"`
					Loudness             string `xml:"loudness,attr"`
					Lra                  string `xml:"lra,attr"`
					Peak                 string `xml:"peak,attr"`
					SamplingRate         string `xml:"samplingRate,attr"`
					StartRamp            string `xml:"startRamp,attr"`
					DisplayTitle         string `xml:"displayTitle,attr"`
					ExtendedDisplayTitle string `xml:"extendedDisplayTitle,attr"`
				} `xml:"Stream"`
			} `xml:"Part"`
		} `xml:"Media"`
	} `xml:"Track"`
}

type MediaContainer struct {
	XMLName   xml.Name `xml:"MediaContainer"`
	Text      string   `xml:",chardata"`
	CommandID string   `xml:"commandID,attr"`
	Timeline  []Track  `xml:"Timeline"`
}

func verifyPlexampConnection() error {
	// todo: verify connection to plexamp
	return nil
}

func verifyPlexConnection() error {
	// todo: verify connection to plex
	return nil
}

func getDefaultThumbnail() []byte {
	if defaultThumbnailData == nil {
		data, err := os.ReadFile("images/plexamp.png")
		if err != nil {
			panic(err)
		}

		defaultThumbnailData = data
	}

	return defaultThumbnailData
}

func getCurrentPlaying() (*MediaContainer, error) {
	// Get the current playing media from Plex
	// /player/timeline/poll?wait=0&includeMetadata=1&commandID=1

	reqUrl := PlexAmpAddress + "/player/timeline/poll?wait=0&includeMetadata=1&commandID=1"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	s := MediaContainer{}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (playing *MediaContainer) getMusicTrack() (*Track, error) {
	for _, track := range playing.Timeline {
		if track.Type == "music" {
			return &track, nil
		}
	}

	return nil, nil
}

func (playing *MediaContainer) getThumbnail() ([]byte, error) {
	// Get the thumbnail of the current playing media
	// /library/metadata/355914/thumb/1724958934
	// $.MediaContainer.Timeline["thumb"]

	track, err := playing.getMusicTrack()
	if err != nil {
		return nil, err
	}

	fmt.Println(track.Track.Thumb)

	if track == nil {
		return nil, nil
	}

	thumbUrl := PlexAddress + track.Track.Thumb
	resp, err := http.Get(thumbUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data[:10]))

	return data, nil
}

func (playing *MediaContainer) getThumbnailFilename() (string, error) {
	track, err := playing.getMusicTrack()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("images/thumb_%s_%s.png", track.RatingKey, cmp.Or(track.Track.UpdatedAt, track.Track.AddedAt))

	return filename, nil
}

func (playing *MediaContainer) writeThumbnail() error {
	filename, err := playing.getThumbnailFilename()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filename); err == nil {
		return nil
	}

	thumbnailData, err := playing.getThumbnail()
	if err != nil {
		return err
	}

	if thumbnailData == nil {
		thumbnailData = getDefaultThumbnail()
	}

	err = os.WriteFile(filename, thumbnailData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func log() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(filepath.Join(wd, "stdout.log"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(f, nil))
	slog.SetDefault(logger)
}

func Handle(event *sdk.ReceivedEvent) error {
	switch event.Event {
	case sdk.DeviceDidConnect:
	case sdk.WillAppear:
		contexts[event.Context] = struct{}{}
	case sdk.TitleParametersDidChange:
	case sdk.KeyDown:
	case sdk.KeyUp:
		// todo: start up plexamp, or show it if it's already running ideally with plexamp:// but that doesn't work right now (only plex://)
	case sdk.WillDisappear:
		delete(contexts, event.Context)
	default:
		slog.Info("", "event", event)
	}

	return nil
}

func main() {
	log()

	// verify connections to Plex and Plexamp
	for _, fn := range []func() error{
		verifyPlexConnection,
		verifyPlexampConnection,
	} {
		if err := fn(); err != nil {
			panic(err)
		}
	}

	// establish connection to streamdeck
	var err error
	if streamdeck, err = sdk.New(); err != nil {
		panic(err)
	}

	slog.Info("", "args", os.Args)

	// Register our handlers
	streamdeck.HandlerFunc(
		Handle,
	)

	// defer cleanup thumbnails
	defer func() {
		files, _ := filepath.Glob("images/thumb_*.png")
		for _, file := range files {
			_ = os.Remove(file)
		}
	}()

	// update the thumbnails + current playback
	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			playing, err := getCurrentPlaying()
			if err != nil {
				continue
			}

			if playing == nil {
				continue
			}

			err = playing.writeThumbnail()
			if err != nil {
				slog.Error("error writing thumbnail", "error", err)

				continue
			}

			filename, err := playing.getThumbnailFilename()
			if err != nil {
				slog.Error("error getting thumbnail filename", "error", err)

				continue
			}

			filename = strings.TrimSuffix(filename, ".png")

			for context := range contexts {
				streamdeck.SetImage(context, filename)
			}
		}
	}()

	slog.Info("", "info", streamdeck.Info)

	// serve the plugin
	streamdeck.Start()
}
