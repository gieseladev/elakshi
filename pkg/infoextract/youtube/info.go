package youtube

import (
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/gieseladev/elakshi/pkg/iso8601"
	"github.com/jinzhu/gorm"
	"google.golang.org/api/youtube/v3"
)

const (
	ytServiceName = "youtube"
)

// TODO use context

func extractHighestResThumbnail(d *youtube.ThumbnailDetails) *youtube.Thumbnail {
	if t := d.Maxres; t != nil {
		return t
	}
	if t := d.Standard; t != nil {
		return t
	}
	if t := d.High; t != nil {
		return t
	}
	if t := d.Medium; t != nil {
		return t
	}

	return d.Default
}

type youtubeExtractor struct {
	db      *gorm.DB
	service *youtube.Service
}

func NewExtractor(db *gorm.DB, service *youtube.Service) *youtubeExtractor {
	return &youtubeExtractor{
		db:      db,
		service: service,
	}
}

func (yt *youtubeExtractor) getChannelByID(id string) (*youtube.Channel, error) {
	result, err := yt.service.Channels.List("snippet").Id(id).Do()
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, errors.New("no channel with the given id")
	}

	return result.Items[0], nil
}

func (yt *youtubeExtractor) getVideoByID(id string) (*youtube.Video, error) {
	result, err := yt.service.Videos.List("contentDetails,snippet").Id(id).Do()
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, errors.New("no video with given id")
	}

	return result.Items[0], nil
}

func (yt *youtubeExtractor) getArtist(video *youtube.Video) (edb.Artist, error) {
	channelID := video.Snippet.ChannelId

	var artist edb.Artist
	found, err := edb.GetModelByExternalRef(yt.db, ytServiceName, channelID, &artist)
	if err != nil {
		return edb.Artist{}, err
	} else if found {
		return artist, nil
	}

	channel, err := yt.getChannelByID(channelID)
	if err != nil {
		return edb.Artist{}, err
	}

	var images []edb.Image
	thumbnail := extractHighestResThumbnail(channel.Snippet.Thumbnails)
	if thumbnail != nil {
		img, err := infoextract.GetImage(yt.db, thumbnail.Url)
		if err != nil {
			return edb.Artist{}, err
		}

		images = append(images, img)
	}

	return edb.Artist{
		Name:               channel.Snippet.Title,
		Images:             images,
		ExternalReferences: []edb.ExternalRef{edb.NewExternalRef(ytServiceName, channelID)},
	}, nil
}

func (yt *youtubeExtractor) trackSourcesFromTracklist(tracklist string, video *youtube.Video) ([]edb.TrackSource, error) {
	return []edb.TrackSource{}, nil
}

func (yt *youtubeExtractor) trackFromVideo(video *youtube.Video) (edb.Track, error) {
	artist, err := yt.getArtist(video)
	if err != nil {
		return edb.Track{}, err
	}

	duration, err := iso8601.ParseDuration(video.ContentDetails.Duration)
	if err != nil {
		return edb.Track{}, err
	}

	var images []edb.Image
	thumbnail := extractHighestResThumbnail(video.Snippet.Thumbnails)
	if thumbnail != nil {
		image, err := infoextract.GetImage(yt.db, thumbnail.Url)
		if err != nil {
			return edb.Track{}, err
		}

		images = append(images, image)
	}

	// TODO use contentDetails.caption to check if a video has captions
	//		and if it does, create a task to parsre them for lyrics!

	return edb.Track{
		Name:     video.Snippet.Title,
		LengthMS: uint32(duration.Milliseconds()),
		Artist:   artist,
		Images:   images,

		ExternalReferences: []edb.ExternalRef{edb.NewExternalRef(ytServiceName, video.Id)},
	}, nil
}

func (yt *youtubeExtractor) parseVideo(video *youtube.Video) (edb.AudioSource, error) {
	videoLength, err := iso8601.ParseDuration(video.ContentDetails.Duration)
	if err != nil {
		return edb.AudioSource{}, err
	}

	var trackSources []edb.TrackSource

	tracklist := infoextract.ExtractTracklistFromText(video.Snippet.Description, videoLength.AsDuration())
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
			EndOffsetMS:   track.LengthMS,
		}}
	}

	return edb.AudioSource{
		Type:         ytServiceName,
		URI:          video.Id,
		TrackSources: trackSources,
	}, nil
}

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
