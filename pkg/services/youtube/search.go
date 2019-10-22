package youtube

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gieseladev/elakshi/pkg/audiosrc"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/iso8601"
	"github.com/gieseladev/elakshi/pkg/lazy"
	"github.com/gieseladev/elakshi/pkg/songtitle"
	"github.com/gieseladev/elakshi/pkg/stringcmp"
	"google.golang.org/api/youtube/v3"
	"html"
	"log"
	"sort"
	"strings"
	"unicode/utf8"
)

// TODO
//		- Interpret track names and search for separate parts. (partially done)

const (
	// minTitleScorePercentage is the required percentage of explained runes in
	// a video title for it to be accepted.
	minTitleScorePercentage = 70
	// durationTolerancePercentage is the allowed duration difference between
	// video and track.
	durationTolerancePercentage = 40
	// minViews is the amount of views (inclusive) a video must have to be
	// considered.
	minViews = 1e2
	// maxViewsForRatioCheck is the max amount of views (inclusive) a video can
	// have before its like ratio is no longer checked.
	maxViewsForRatioCheck = 1e4
	// minLikeRatio is the required like ratio (inclusive) for a video which
	// has less than maxViewsForRatioCheck views.
	minLikeRatio = 50
)

func (yt *youtubeService) getSearchResultsByQuery(ctx context.Context, query string) ([]*youtube.SearchResult, error) {
	resp, err := yt.service.Search.
		List("snippet").
		Context(ctx).
		Type("video").
		SafeSearch("none").
		MaxResults(10).
		Q(query).Do()
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (yt *youtubeService) getVideosByID(ctx context.Context, videoIDs ...string) ([]*youtube.Video, error) {
	resp, err := yt.service.Videos.
		List("contentDetails,statistics").
		Context(ctx).
		Id(strings.Join(videoIDs, ",")).
		Do()
	if err != nil {
		return nil, err
	}

	// FIXME the following checks are only for testing/debugging purposes
	if len(resp.Items) != len(videoIDs) {
		panic("not enough videos returned")
	}

	for i, item := range resp.Items {
		if item.Id != videoIDs[i] {
			panic("youtube video order expectation not met!")
		}
	}

	return resp.Items, nil
}

type scoredResult struct {
	TitleScore      int // percentage of explainable runes in the title.
	VideoID         string
	Result          *youtube.SearchResult
	Video           *youtube.Video
	VideoDurationMS uint64
}

func strContainedSurroundedInAnyOfL(substr string, strLs ...lazy.StringL) bool {
	for _, strL := range strLs {
		if stringcmp.ContainsSurrounded(strL(), substr) {
			return true
		}
	}

	return false
}

func scoreSearchResult(track edb.Track, sr *youtube.SearchResult) (scoredResult, bool) {
	res := scoredResult{
		VideoID: sr.Id.VideoId,
		Result:  sr,
	}

	snippet := sr.Snippet
	// we don't want to accept live streams here.
	if snippet.LiveBroadcastContent != "none" {
		return res, false
	}

	// TODO interpret buzzwords in titles. Let's create a separate library
	//      for this, since it seems rather common.
	//	    especially ["official video", "original mix", "original"]
	//	    and probably more are implied when they're absent. Hence they
	//		ought to be removed from video titles.
	//		If they're also part of the track title, they need to be
	//		removed from it as well.
	//		"ft." / "feat." ought to be detected as well.

	// YouTube deliberately returns html escaped strings.
	// See: https://issuetracker.google.com/u/1/issues/128673539
	videoTitle := html.UnescapeString(snippet.Title)

	cleanVideoTitle := stringcmp.GetWordsFocusedString(videoTitle)

	// TODO extract
	cleanTrackName := stringcmp.GetWordsFocusedString(track.Name)

	cleanChannelTitleL := lazy.StringFunc(func() string {
		return stringcmp.GetWordsFocusedString(
			html.UnescapeString(snippet.ChannelTitle))
	})

	cleanDescriptionL := lazy.StringFunc(func() string {
		return stringcmp.GetWordsFocusedString(
			html.UnescapeString(snippet.Description))
	})

	// We track the amount of runes we can reasonably explain the presence
	// of in the title. We can then use the percentage of "explainable"
	// runes to score a result.
	explainedRunes := 0

	containedSurroundedInAny := func(s string) bool {
		s = stringcmp.GetWordsFocusedString(s)

		switch {
		case stringcmp.ContainsSurrounded(cleanVideoTitle, s):
			explainedRunes += utf8.RuneCountInString(s)
		// the title might contain additional artists which may be part of
		// the channel title.
		// Other parts can be part of the description as well.
		case strContainedSurroundedInAnyOfL(s,
			cleanDescriptionL, cleanChannelTitleL):
		default:
			// if we can't find all parts, reject result
			return false
		}

		return true
	}

	// if the video title doesn't contain the track's name ignore it!
	if stringcmp.ContainsSurroundedIgnoreSpace(cleanVideoTitle, cleanTrackName) {
		explainedRunes += utf8.RuneCountInString(cleanTrackName)
	} else {
		// TODO extract
		trackTitle := songtitle.ParseTitle(track.Name)

		for _, part := range trackTitle.BaselineParts {
			part = stringcmp.GetWordsFocusedString(part)

			if !stringcmp.ContainsSurrounded(cleanVideoTitle, part) {
				return res, false
			}

			explainedRunes += utf8.RuneCountInString(part)
		}

		for _, part := range trackTitle.GuestAppearances {
			if !containedSurroundedInAny(part) {
				return res, false
			}
		}

		for _, part := range trackTitle.OtherParts {
			if !containedSurroundedInAny(part) {
				return res, false
			}
		}
	}

	if track.ArtistID != nil {
		if !containedSurroundedInAny(track.Artist.Name) {
			// if we can't find the artist, ignore the result entirely!
			return res, false
		}
	} else {
		// TODO what if no artist?
	}

	// Check for multiple artists and remove them from title.
	foundMultipleArtists := false
	for _, artist := range track.AdditionalArtists { // TODO extract
		cleanArtistName := stringcmp.GetWordsFocusedString(artist.Name)
		// we're not searching the description / channel title here because
		// we don't actually care whether the additional artists are there or
		// not. We're just trying to explain the title!
		if stringcmp.ContainsSurrounded(cleanVideoTitle, cleanArtistName) {
			// allow a space between artist and title
			explainedRunes += utf8.RuneCountInString(cleanArtistName) + 1
			foundMultipleArtists = true
		}
	}

	// if we found multiple artists we can explain the "ft.".
	if foundMultipleArtists {
		// TODO use songtitle library for this
		word := stringcmp.ContainsAnyOf(videoTitle, "feat.", "ft.", "featuring")
		if word != "" {
			// allow 2 spaces to the side of the "ft."
			explainedRunes += utf8.RuneCountInString(word) + 2
		}
	}

	// explain the album name in the title.
	if track.AlbumID != nil {
		// TODO extract
		cleanAlbumName := stringcmp.GetWordsFocusedString(track.Album.Name)
		if stringcmp.ContainsSurrounded(cleanVideoTitle, cleanAlbumName) {
			// allow 1 space
			explainedRunes += utf8.RuneCountInString(cleanAlbumName) + 1
		}
	}

	// TODO extract rune count of cleanVideoTitle
	res.TitleScore = 100 * explainedRunes / utf8.RuneCountInString(cleanVideoTitle)
	if res.TitleScore <= minTitleScorePercentage {
		// TODO check whether there are any filler or content labels in the
		//  string we can ignore to get the score up.
		return res, false
	}

	return res, true
}

func (yt *youtubeService) basicSearch(ctx context.Context, track edb.Track) ([]scoredResult, error) {
	var query string
	if track.ArtistID != nil {
		query = track.Artist.Name + " - " + track.Name
	} else {
		query = track.Name
	}

	searchResults, err := yt.getSearchResultsByQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	var relevantResults []scoredResult
	for _, sr := range searchResults {
		if res, ok := scoreSearchResult(track, sr); ok {
			relevantResults = append(relevantResults, res)
		}
	}

	if len(relevantResults) == 0 {
		return nil, nil
	}

	// stable sort to keep youtube's order in tact.
	sort.SliceStable(relevantResults, func(i, j int) bool {
		return relevantResults[i].TitleScore > relevantResults[j].TitleScore
	})

	return relevantResults, nil
}

func (yt *youtubeService) buildResults(track edb.Track, relevantResults []scoredResult) []audiosrc.Result {
	results := make([]audiosrc.Result, len(relevantResults))

	for i, res := range relevantResults {
		source := edb.TrackSource{
			TrackID:       track.ID,
			StartOffsetMS: 0,
			EndOffsetMS:   uint32(res.VideoDurationMS),

			AudioSource: &edb.AudioSource{
				Type: ytServiceName,
				URI:  res.VideoID,
			},
		}

		results[i] = audiosrc.Result{TrackSource: source}
	}

	return results
}

func (yt *youtubeService) Search(ctx context.Context, track edb.Track) ([]audiosrc.Result, error) {
	// TODO use external references to find youtube video

	relevantResults, err := yt.basicSearch(ctx, track)
	if err != nil {
		return nil, err
	}

	if len(relevantResults) == 0 {
		return nil, nil
	}

	videoIDs := make([]string, len(relevantResults))
	for i, res := range relevantResults {
		videoIDs[i] = res.VideoID
	}

	videos, err := yt.getVideosByID(ctx, videoIDs...)
	if err != nil {
		return nil, err
	}

	prevResults := relevantResults
	relevantResults = relevantResults[:0]
	for i, video := range videos {
		result := prevResults[i]
		result.Video = video

		videoDuration, err := iso8601.ParseDuration(video.ContentDetails.Duration)
		if err != nil {
			log.Println("couldn't parse video duration:", video.ContentDetails.Duration, "error:", err)
			sentry.CaptureException(err)
			continue
		}
		result.VideoDurationMS = videoDuration.Milliseconds()

		// remove videos whose durations don't resemble the targeted one.
		if !durationWithinPercentage(uint64(track.LengthMS), result.VideoDurationMS, durationTolerancePercentage) {
			continue
		}

		stats := video.Statistics
		if stats.ViewCount < minViews {
			continue
		}

		if stats.ViewCount <= maxViewsForRatioCheck && getLikeRatio(stats) < minLikeRatio {
			continue
		}

		relevantResults = append(relevantResults, result)
	}

	return yt.buildResults(track, relevantResults), nil
}

// durationWithinPercentage checks whether the actual value is within perc
// percent of the expected value.
// perc is the percentage in centi.
func durationWithinPercentage(expected, actual uint64, perc uint64) bool {
	// this function is definitely over-optimised considering we're using
	// unicode string searches and calling multiple APIs in the process.
	// But hey, you gotta do what you can!
	var diff uint64
	if expected > actual {
		diff = expected - actual
	} else {
		diff = actual - expected
	}

	return 100*diff <= perc*expected
}

func getLikeRatio(stats *youtube.VideoStatistics) uint {
	total := stats.LikeCount + stats.DislikeCount
	if total == 0 {
		return 100
	}

	return uint(100 * stats.LikeCount / total)
}
