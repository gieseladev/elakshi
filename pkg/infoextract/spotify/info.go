package spotify

import (
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/errutils"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"sync"
)

const (
	spotifyServiceName = "spotify"
)

type spotifyExtractor struct {
	client *spotify.Client
	db     *gorm.DB
}

func NewExtractor(client *spotify.Client) *spotifyExtractor {
	return &spotifyExtractor{client: client}
}

func (s *spotifyExtractor) genresFromGenres(genres []string) ([]edb.Genre, error) {
	gens := make([]edb.Genre, len(genres))

	for i, genre := range genres {
		gens[i] = edb.Genre{
			Name: genre,
		}
	}

	return gens, nil
}

func (s *spotifyExtractor) imageFromImage(image spotify.Image) (edb.Image, error) {
	return edb.Image{
		// TODO download images
		URI: image.URL,
	}, nil
}

func (s *spotifyExtractor) imagesFromImages(images []spotify.Image) ([]edb.Image, error) {
	// TODO Are there even multiple different images?
	// 		Or is it only ever 1 image in different resolutions

	var mux sync.Mutex
	var wg sync.WaitGroup
	var errs errutils.MultiError

	imgs := make([]edb.Image, 0, len(images))
	wg.Add(len(images))
	for _, image := range images {
		go func(image spotify.Image) {
			defer wg.Done()
			i, err := s.imageFromImage(image)

			mux.Lock()
			defer mux.Unlock()

			if err == nil {
				imgs = append(imgs, i)
			} else {
				errs = append(errs, err)
			}
		}(image)
	}

	wg.Wait()

	return imgs, errs.AsError()
}

func (s *spotifyExtractor) artistFromFullArtist(artist *spotify.FullArtist) (edb.Artist, error) {
	images, err := s.imagesFromImages(artist.Images)
	if err != nil {
		return edb.Artist{}, err
	}

	genres, err := s.genresFromGenres(artist.Genres)
	if err != nil {
		return edb.Artist{}, err
	}

	return edb.Artist{
		Name:   artist.Name,
		Images: images,
		Genres: genres,
	}, nil
}

func (s *spotifyExtractor) artistsFromFullArtists(artists []*spotify.FullArtist) ([]edb.Artist, error) {
	var mux sync.Mutex
	var wg sync.WaitGroup
	var errs errutils.MultiError

	rawArts := make([]edb.Artist, len(artists))

	wg.Add(len(artists))
	for i, artist := range artists {
		go func(i int, artist *spotify.FullArtist) {
			defer wg.Done()

			a, err := s.artistFromFullArtist(artist)
			if err == nil {
				rawArts[i] = a
			} else {
				mux.Lock()
				errs = append(errs, err)
				mux.Unlock()
			}

		}(i, artist)
	}

	wg.Wait()

	err := errs.AsError()
	if err != nil {
		return nil, err
	}

	arts := rawArts[:0]
	for _, a := range rawArts {
		if !a.IsEmpty() {
			arts = append(arts, a)
		}
	}

	return arts, nil
}

func (s *spotifyExtractor) artistsFromSimpleArtists(artists []spotify.SimpleArtist) ([]edb.Artist, error) {
	artistIDs := make([]spotify.ID, len(artists))
	for i, artist := range artists {
		artistIDs[i] = artist.ID
	}

	fullArtists, err := s.client.GetArtists(artistIDs...)
	if err != nil {
		return nil, err
	}

	return s.artistsFromFullArtists(fullArtists)
}

func (s *spotifyExtractor) albumFromFullAlbum(album *spotify.FullAlbum) (edb.Album, error) {
	artists, err := s.artistsFromSimpleArtists(album.Artists)
	if err != nil {
		return edb.Album{}, err
	}

	images, err := s.imagesFromImages(album.Images)
	if err != nil {
		return edb.Album{}, err
	}

	genres, err := s.genresFromGenres(album.Genres)
	if err != nil {
		return edb.Album{}, err
	}

	return edb.Album{
		Name:        album.Name,
		Images:      images,
		Artists:     artists,
		ReleaseDate: album.ReleaseDateTime(),
		Genres:      genres,
	}, nil
}

func (s *spotifyExtractor) albumFromSimpleAlbum(album spotify.SimpleAlbum) (edb.Album, error) {
	fullAlbum, err := s.client.GetAlbum(album.ID)
	if err != nil {
		return edb.Album{}, err
	}

	return s.albumFromFullAlbum(fullAlbum)
}

func (s *spotifyExtractor) trackFromFullTrack(track *spotify.FullTrack) (edb.Track, error) {
	artists, err := s.artistsFromSimpleArtists(track.Artists)
	if err != nil {
		return edb.Track{}, err
	}

	var mainArtist edb.Artist
	if len(artists) > 0 {
		mainArtist = artists[0]
		artists = artists[1:]
	}

	album, err := s.albumFromSimpleAlbum(track.Album)
	if err != nil {
		return edb.Track{}, err
	}

	if track.Duration < 0 {
		return edb.Track{}, errors.New("spotify provided negative duration")
	}

	// TODO external ids

	return edb.Track{
		Name:              track.Name,
		Artist:            mainArtist,
		AdditionalArtists: artists,
		Album:             album,
		LengthMS:          uint32(track.Duration),
	}, nil
}

func (s *spotifyExtractor) trackFromTrackID(trackID string) (edb.Track, error) {
	var t edb.Track
	err := s.db.Scopes(edb.GetModelByExternalRef(&t, spotifyServiceName, trackID)).Scan(&t).Error
	if err == nil {
		return t, nil
	} else if !gorm.IsRecordNotFoundError(err) {
		return edb.Track{}, err
	}

	track, err := s.client.GetTrack(spotify.ID(trackID))
	if err != nil {
		return edb.Track{}, err
	}

	return s.trackFromFullTrack(track)
}
