package youtube

import (
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract/common"
	"google.golang.org/api/youtube/v3"
	"strconv"
	"strings"
	"time"
)

const (
	ExtractorType = "youtube"
)

type youtubeExtractor struct {
	service *youtube.Service
}

func (yt *youtubeExtractor) getVideoByID(id string) (*youtube.Video, error) {
	result, err := yt.service.Videos.List("contentDetails,snippet").Id(id).Do()
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		// items mustn't be empty, the api should return an error if the id
		// doesn't exist.
		panic("no results from youtube")
	}

	return result.Items[0], nil
}

func parseYTDuration(duration string) (time.Duration, error) {
	errInvalidFormat := errors.New("invalid format")

	// len(PTnMnS) = 6
	if len(duration) < 6 {
		return 0, errInvalidFormat
	}

	// skip PT and remove S
	duration = duration[2 : len(duration)-1]

	mIndex := strings.IndexByte(duration, 'M')
	if mIndex < 0 {
		return 0, errInvalidFormat
	}

	minutes, err := strconv.ParseInt(duration[:mIndex], 10, 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseInt(duration[mIndex+1:], 10, 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

func (yt *youtubeExtractor) parseVideo(video *youtube.Video) (edb.AudioSource, error) {
	if video.Snippet == nil || video.ContentDetails == nil {
		panic("video requires snippet, contentDetails")
	}

	videoLength, err := parseYTDuration(video.ContentDetails.Duration)
	if err != nil {
		panic("couldn't parse iso8601 duration")
	}

	var trackSources []edb.TrackSource

	tracklist := common.ExtractTracklistFromText(video.Snippet.Description, videoLength)
	if len(tracklist) > 1 {
		// TODO handle tracklist
	} else {

		// TODO search artist by youtube channel
		trackSources = []edb.TrackSource{{
			// TODO create track
			Track:         edb.Track{},
			StartOffsetMS: 0,
			EndOffsetMS:   uint32(videoLength.Milliseconds()),
		}}
	}

	return edb.AudioSource{
		Type:         ExtractorType,
		URI:          video.Id,
		TrackSources: trackSources,
	}, nil
}
