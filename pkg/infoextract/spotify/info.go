package spotify

import (
	"context"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
)

const (
	spotifyServiceName = "spotify"
)

// TODO use other external references (other than spotify id) to search for
//  	entities.

// TODO tracks should also resolve cross references in the form of title - artist
// 		or title - album.

type spotifyExtractor struct {
	db     *gorm.DB
	client *spotify.Client
}

func NewExtractor(db *gorm.DB, client *spotify.Client) *spotifyExtractor {
	return &spotifyExtractor{db: db, client: client}
}

func (s *spotifyExtractor) extRefsFromIDs(externalIDs map[string]string) []edb.ExternalRef {
	var refs []edb.ExternalRef

	if isrc, ok := externalIDs["isrc"]; ok {
		refs = append(refs, edb.NewExternalRef("isrc", isrc))
	}

	if ean, ok := externalIDs["ean"]; ok {
		refs = append(refs, edb.NewExternalRef("ean", ean))
	}

	if upc, ok := externalIDs["upc"]; ok {
		refs = append(refs, edb.NewExternalRef("upc", upc))
	}

	return refs
}

// TODO move to common
func (s *spotifyExtractor) GetImage(uri string) (edb.Image, error) {
	image := edb.Image{
		SourceURI: uri,
	}

	err := s.db.FirstOrCreate(&image, &image).Error
	if err != nil {
		return edb.Image{}, err
	}

	// TODO schedule download if URI is nil

	return image, nil
}

func (s *spotifyExtractor) imagesFromImages(images []spotify.Image) ([]edb.Image, error) {
	if len(images) == 0 {
		return nil, nil
	}

	// the first image will always be the "widest"
	i, err := s.GetImage(images[0].URL)
	if err != nil {
		return nil, err
	}

	return []edb.Image{i}, nil
}

func (s *spotifyExtractor) artistFromFullArtist(artist *spotify.FullArtist) (edb.Artist, error) {
	images, err := s.imagesFromImages(artist.Images)
	if err != nil {
		return edb.Artist{}, err
	}

	genres, err := infoextract.GetGenresByName(s.db, artist.Genres...)
	if err != nil {
		return edb.Artist{}, err
	}

	extRefs := make([]edb.ExternalRef, 0)
	extRefs = append(extRefs, edb.NewExternalRef(spotifyServiceName, string(artist.ID)))

	return edb.Artist{
		Name:   artist.Name,
		Images: images,
		Genres: genres,

		ExternalReferences: extRefs,
	}, nil
}

func (s *spotifyExtractor) GetArtists(ctx context.Context, ids ...string) ([]edb.Artist, error) {
	artists := make([]edb.Artist, len(ids))
	// this isn't the most efficient way to do this, but we're mostly dealing
	// with 1-2 artists so it's not all that bad.
	for i, id := range ids {
		var artist edb.Artist
		found, err := edb.GetModelByExternalRef(s.db, spotifyServiceName, id, &artist)
		if err != nil {
			return nil, err
		}

		if !found {
			a, err := s.client.GetArtist(spotify.ID(id))
			if err != nil {
				return nil, err
			}
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			artist, err = s.artistFromFullArtist(a)
			if err != nil {
				return nil, err
			}

			err = s.db.Create(&artist).Error
			if err != nil {
				return nil, err
			}
		}

		artists[i] = artist
	}

	return artists, nil
}

func (s *spotifyExtractor) getArtistsFromSimpleArtists(ctx context.Context, artists []spotify.SimpleArtist) ([]edb.Artist, error) {
	artistIDs := make([]string, len(artists))
	for i, artist := range artists {
		artistIDs[i] = string(artist.ID)
	}

	return s.GetArtists(ctx, artistIDs...)
}

func (s *spotifyExtractor) albumFromFullAlbum(ctx context.Context, album *spotify.FullAlbum) (edb.Album, error) {
	artists, err := s.getArtistsFromSimpleArtists(ctx, album.Artists)
	if err != nil {
		return edb.Album{}, err
	}

	images, err := s.imagesFromImages(album.Images)
	if err != nil {
		return edb.Album{}, err
	}

	genres, err := infoextract.GetGenresByName(s.db, album.Genres...)
	if err != nil {
		return edb.Album{}, err
	}

	releaseDate := album.ReleaseDateTime()

	extRefs := s.extRefsFromIDs(album.ExternalIDs)
	extRefs = append(extRefs, edb.NewExternalRef(spotifyServiceName, string(album.ID)))

	return edb.Album{
		Name:        album.Name,
		Images:      images,
		Artists:     artists,
		ReleaseDate: &releaseDate,
		Genres:      genres,

		ExternalReferences: extRefs,
	}, nil
}

func (s *spotifyExtractor) GetAlbum(ctx context.Context, albumID string) (edb.Album, error) {
	var a edb.Album
	found, err := edb.GetModelByExternalRef(s.db, spotifyServiceName, albumID, &a)
	if err != nil {
		return edb.Album{}, err
	} else if found {
		return a, nil
	}

	album, err := s.client.GetAlbum(spotify.ID(albumID))
	if err != nil {
		return edb.Album{}, err
	}
	if err := ctx.Err(); err != nil {
		return edb.Album{}, err
	}

	a, err = s.albumFromFullAlbum(ctx, album)
	if err != nil {
		return edb.Album{}, err
	}

	err = s.db.Create(&a).Error
	return a, err
}

func (s *spotifyExtractor) trackFromFullTrack(ctx context.Context, track *spotify.FullTrack) (edb.Track, error) {
	artists, err := s.getArtistsFromSimpleArtists(ctx, track.Artists)
	if err != nil {
		return edb.Track{}, err
	}

	var mainArtist edb.Artist
	if len(artists) > 0 {
		mainArtist = artists[0]
		artists = artists[1:]
	}

	album, err := s.GetAlbum(ctx, string(track.Album.ID))
	if err != nil {
		return edb.Track{}, err
	}

	if track.Duration < 0 {
		return edb.Track{}, errors.New("spotify provided negative duration")
	}

	extRefs := s.extRefsFromIDs(track.ExternalIDs)
	extRefs = append(extRefs, edb.NewExternalRef(spotifyServiceName, string(track.ID)))

	return edb.Track{
		Name:              track.Name,
		Artist:            mainArtist,
		AdditionalArtists: artists,
		Album:             album,
		LengthMS:          uint32(track.Duration),

		ExternalReferences: extRefs,
	}, nil
}

func (s *spotifyExtractor) GetTrack(ctx context.Context, trackID string) (edb.Track, error) {
	var t edb.Track
	found, err := edb.GetModelByExternalRef(s.db, spotifyServiceName, trackID, &t)
	if err != nil {
		return edb.Track{}, err
	} else if found {
		return t, nil
	}

	track, err := s.client.GetTrack(spotify.ID(trackID))
	if err != nil {
		return edb.Track{}, err
	}
	if err := ctx.Err(); err != nil {
		return edb.Track{}, err
	}

	// find by other external ids
	found, err = edb.GetModelByExternalRefs(s.db, &t, s.extRefsFromIDs(track.ExternalIDs))
	if err != nil {
		return edb.Track{}, err
	} else if found {
		// create external reference for spotify
		err := s.db.Model(&t).
			Association("ExternalReferences").
			Append(edb.NewExternalRef(spotifyServiceName, string(track.ID))).
			Error
		if err != nil {
			return edb.Track{}, err
		}

		return t, nil
	}

	// create new track

	t, err = s.trackFromFullTrack(ctx, track)
	if err != nil {
		return edb.Track{}, err
	}

	err = s.db.Create(&t).Error
	if err != nil {
		return edb.Track{}, err
	}

	return t, nil
}
