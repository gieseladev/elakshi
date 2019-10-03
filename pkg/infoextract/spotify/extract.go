package spotify

import (
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/errutils"
	"github.com/gieseladev/elakshi/pkg/infoextract/common"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"sync"
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

func (s *spotifyExtractor) imageFromImage(image spotify.Image) (edb.Image, error) {
	return edb.Image{
		// TODO schedule download
		URI: image.URL,
	}, nil
}

func (s *spotifyExtractor) imagesFromImages(images []spotify.Image) ([]edb.Image, error) {
	// TODO It seems like there's only ever one image (but in varying sizes)!

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

	genres, err := common.GetGenresByName(s.db, artist.Genres...)
	if err != nil {
		return edb.Artist{}, err
	}

	extRefs := []edb.ExternalRef{edb.NewExternalRef(spotifyServiceName, string(artist.ID))}

	return edb.Artist{
		Name:   artist.Name,
		Images: images,
		Genres: genres,

		ExternalReferences: extRefs,
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

func (s *spotifyExtractor) GetArtists(ids ...string) ([]edb.Artist, error) {
	artists := make([]edb.Artist, len(ids))
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

func (s *spotifyExtractor) artistsFromSimpleArtists(artists []spotify.SimpleArtist) ([]edb.Artist, error) {
	artistIDs := make([]string, len(artists))
	for i, artist := range artists {
		artistIDs[i] = string(artist.ID)
	}

	return s.GetArtists(artistIDs...)
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

	genres, err := common.GetGenresByName(s.db, album.Genres...)
	if err != nil {
		return edb.Album{}, err
	}

	releaseDate := album.ReleaseDateTime()

	extRefs := []edb.ExternalRef{
		edb.NewExternalRef(spotifyServiceName, string(album.ID)),
	}

	return edb.Album{
		Name:        album.Name,
		Images:      images,
		Artists:     artists,
		ReleaseDate: &releaseDate,
		Genres:      genres,

		ExternalReferences: extRefs,
	}, nil
}

func (s *spotifyExtractor) GetAlbum(albumID string) (edb.Album, error) {
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

	a, err = s.albumFromFullAlbum(album)
	if err != nil {
		return edb.Album{}, err
	}

	err = s.db.Create(&a).Error
	return a, err
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

	album, err := s.GetAlbum(string(track.Album.ID))
	if err != nil {
		return edb.Track{}, err
	}

	if track.Duration < 0 {
		return edb.Track{}, errors.New("spotify provided negative duration")
	}

	extRefs := []edb.ExternalRef{edb.NewExternalRef(spotifyServiceName, string(track.ID))}

	return edb.Track{
		Name:              track.Name,
		Artist:            mainArtist,
		AdditionalArtists: artists,
		Album:             album,
		LengthMS:          uint32(track.Duration),

		ExternalReferences: extRefs,
	}, nil
}

func (s *spotifyExtractor) GetTrack(trackID string) (edb.Track, error) {
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

	t, err = s.trackFromFullTrack(track)
	if err != nil {
		return edb.Track{}, err
	}

	err = s.db.Create(&t).Error
	if err != nil {
		return edb.Track{}, err
	}

	return t, nil
}
