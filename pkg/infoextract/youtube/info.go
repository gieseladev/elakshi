package youtube

import (
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract/common"
	"github.com/gieseladev/elakshi/pkg/iso8601"
	"github.com/jinzhu/gorm"
	"google.golang.org/api/youtube/v3"
)

const (
	ytServiceName = "youtube"
)

type youtubeExtractor struct {
	db      *gorm.DB
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

func (yt *youtubeExtractor) trackSourcesFromTracklist(tracklist string, video *youtube.Video) ([]edb.TrackSource, error) {
	return []edb.TrackSource{}, nil
}

func (yt *youtubeExtractor) trackFromVideo(video *youtube.Video) (edb.Track, error) {
	// TODO map to database
	// TODO search artist by youtube channel
	return edb.Track{
		Name: video.Snippet.Title,
		AdditionalArtists: []edb.Artist{{
			Name: video.Snippet.ChannelTitle,
		}},
	}, nil
}

func (yt *youtubeExtractor) parseVideo(video *youtube.Video) (edb.AudioSource, error) {
	if video.Snippet == nil || video.ContentDetails == nil {
		panic("video requires snippet, contentDetails")
	}

	videoLength, err := iso8601.ParseDuration(video.ContentDetails.Duration)
	if err != nil {
		panic("couldn't parse iso8601 duration")
	}

	var trackSources []edb.TrackSource

	tracklist := common.ExtractTracklistFromText(video.Snippet.Description, videoLength.AsDuration())
	if len(tracklist) > 1 {
		// TODO pass tracklist
		trackSources, err = yt.trackSourcesFromTracklist("", video)
		if err != nil {
			return edb.AudioSource{}, nil
		}
	} else {
		track, err := yt.trackFromVideo(video)
		if err != nil {
			return edb.AudioSource{}, nil
		}

		trackSources = []edb.TrackSource{{
			Track:         track,
			StartOffsetMS: 0,
			EndOffsetMS:   uint32(videoLength.Milliseconds()),
		}}
	}

	return edb.AudioSource{
		Type:         ytServiceName,
		URI:          video.Id,
		TrackSources: trackSources,
	}, nil
}

// TODO use contentDetails.caption to check if a video has captions

// TODO only use thumbnails as images when contentDetails.hasCustomThumbnail is
//  true

func tracksFromAudioSource(audio edb.AudioSource) []edb.Track {
	tracks := make([]edb.Track, len(audio.TrackSources))
	for i, source := range audio.TrackSources {
		tracks[i] = source.Track
	}

	return tracks
}

func (yt *youtubeExtractor) GetTracks(videoID string) ([]edb.Track, error) {
	var track edb.Track
	found, err := edb.GetModelByExternalRef(yt.db, ytServiceName, videoID, &track)
	if err != nil {
		return nil, err
	} else if found {
		return []edb.Track{track}, nil
	}

	// find by assigned audio source
	var audioSource edb.AudioSource
	err = yt.db.
		Preload("TrackSources.Track").
		Take(&audioSource, "type = ? AND uri = ?", ytServiceName, videoID).
		Error
	if err == nil {
		return tracksFromAudioSource(audioSource), nil
	} else if !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	video, err := yt.getVideoByID(videoID)
	if err != nil {
		return nil, err
	}

	audioSource, err = yt.parseVideo(video)
	if err != nil {
		return nil, err
	}

	if err := yt.db.Create(&audioSource).Error; err != nil {
		return nil, err
	}

	return tracksFromAudioSource(audioSource), nil
}
