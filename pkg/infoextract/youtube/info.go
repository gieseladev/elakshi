package youtube

import (
	"context"
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

func tracksFromAudioSource(audio edb.AudioSource) []edb.Track {
	tracks := make([]edb.Track, len(audio.TrackSources))
	for i, source := range audio.TrackSources {
		tracks[i] = source.Track
	}

	return tracks
}

type youtubeExtractor struct {
	db      *gorm.DB
	service *youtube.Service
}

var (
	ErrIDInvalid = errors.New("id invalid")
)

func (yt *youtubeExtractor) getChannelByID(ctx context.Context, id string) (*youtube.Channel, error) {
	result, err := yt.service.Channels.
		List("snippet").
		Id(id).
		Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, ErrIDInvalid
	}

	return result.Items[0], nil
}

func (yt *youtubeExtractor) getVideoByID(ctx context.Context, id string) (*youtube.Video, error) {
	result, err := yt.service.Videos.
		List("contentDetails,snippet").
		Id(id).
		Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, ErrIDInvalid
	}

	return result.Items[0], nil
}

func (yt *youtubeExtractor) getArtist(ctx context.Context, video *youtube.Video) (edb.Artist, error) {
	channelID := video.Snippet.ChannelId

	// TODO don't use artist if channel name isn't contained in the video!

	var artist edb.Artist
	found, err := edb.GetModelByExternalRef(yt.db, ytServiceName, channelID, &artist)
	if err != nil {
		return edb.Artist{}, err
	} else if found {
		return artist, nil
	}

	channel, err := yt.getChannelByID(ctx, channelID)
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

func (yt *youtubeExtractor) trackSourcesFromTracklist(tracklist interface{}, video *youtube.Video) ([]edb.TrackSource, error) {
	return []edb.TrackSource{}, nil
}

func (yt *youtubeExtractor) trackFromVideo(ctx context.Context, video *youtube.Video) (edb.Track, error) {
	artist, err := yt.getArtist(ctx, video)
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

func (yt *youtubeExtractor) parseVideo(ctx context.Context, video *youtube.Video) (edb.AudioSource, error) {
	videoLength, err := iso8601.ParseDuration(video.ContentDetails.Duration)
	if err != nil {
		return edb.AudioSource{}, err
	}

	var trackSources []edb.TrackSource

	// TODO get tracklist from library!
	var tracklist []interface{}
	if len(tracklist) > 1 {
		// TODO pass tracklist
		_ = videoLength // let's keep videoLength for now
		trackSources, err = yt.trackSourcesFromTracklist(nil, video)
		if err != nil {
			return edb.AudioSource{}, nil
		}
	} else {
		track, err := yt.trackFromVideo(ctx, video)
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

func (yt *youtubeExtractor) GetTracks(ctx context.Context, videoID string) ([]edb.Track, error) {
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

	video, err := yt.getVideoByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	audioSource, err = yt.parseVideo(ctx, video)
	if err != nil {
		return nil, err
	}

	if err := yt.db.Create(&audioSource).Error; err != nil {
		return nil, err
	}

	return tracksFromAudioSource(audioSource), nil
}

func NewExtractor(db *gorm.DB, service *youtube.Service) *youtubeExtractor {
	return &youtubeExtractor{
		db:      db,
		service: service,
	}
}
