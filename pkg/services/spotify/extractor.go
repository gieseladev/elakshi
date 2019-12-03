package spotify

import (
	"context"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/zmb3/spotify"
)

func (s *spotifyService) Extract(ctx context.Context, uri string) (interface{}, error) {
	typ, id, err := parseURI(uri)
	if err == ErrInvalidSpotifyURI {
		return nil, infoextract.ErrURIInvalid
	} else if err != nil {
		return nil, err
	}

	switch typ {
	case "track":
		// TODO handle spotify 404 errors and resolve to uri invalid or something
		return s.GetTrack(ctx, id)
	default:
		// TODO global invalid "type" error
		return nil, infoextract.ErrURIInvalid
	}
}

func (s *spotifyService) extRefsFromIDs(externalIDs map[string]string) []edb.ExternalRef {
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

func (s *spotifyService) imagesFromImages(images []spotify.Image) ([]edb.Image, error) {
	if len(images) == 0 {
		return nil, nil
	}

	// the first image will always be the "widest"
	i, err := infoextract.GetImage(s.db, images[0].URL)
	if err != nil {
		return nil, err
	}

	return []edb.Image{i}, nil
}

func (s *spotifyService) artistFromFullArtist(artist *spotify.FullArtist) (edb.Artist, error) {
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

func (s *spotifyService) GetArtists(ctx context.Context, ids ...string) ([]edb.Artist, error) {
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

func (s *spotifyService) getArtistsFromSimpleArtists(ctx context.Context, artists []spotify.SimpleArtist) ([]edb.Artist, error) {
	artistIDs := make([]string, len(artists))
	for i, artist := range artists {
		artistIDs[i] = string(artist.ID)
	}

	return s.GetArtists(ctx, artistIDs...)
}

func (s *spotifyService) albumFromFullAlbum(ctx context.Context, album *spotify.FullAlbum) (edb.Album, error) {
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

func (s *spotifyService) GetAlbum(ctx context.Context, albumID string) (edb.Album, error) {
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

func (s *spotifyService) trackFromFullTrack(ctx context.Context, track *spotify.FullTrack) (edb.Track, error) {
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

func (s *spotifyService) GetTrack(ctx context.Context, trackID string) (edb.Track, error) {
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
