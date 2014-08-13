// Copyright 2014, Amahi.  All f reserved.
// Use of this source code is governed by the
// license that can be found in the LICENSE file.

// Functions for getting TV metadata from TVRage
// See this page for detail: http://services.tvrage.com/info.php?page=main
package tvrage

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
)

type TVRage struct {
	tv_rage_api_key string
	tv_db_api_key string
}

type tvrageResult struct {
	XMLName     xml.Name     `xml:"Results"`
	ShowDetails []tvrageShow `xml:"show"`
}

type tvrageShow struct {
	Id   int    `xml:"showid"`
	Name string `xml:"name"`
}

type tvdbResult struct {
	XMLName       xml.Name      `xml:"Data"`
	SeriesDetails []tvdbDetails `xml:"Series"`
}

type tvdbDetails struct {
	SeriesId string `xml:"seriesid"`
	Language string `xml:"language"`
	Name     string `xml:"SeriesName"`
}

type tvMetadata struct {
	Media_type string
	SeriesName string `xml:"Series>SeriesName"`
	Banner_Url string
	Actors     string `xml:"Series>Actors"`
	Overview   string `xml:"Series>Overview"`
	Banner     string `xml:"Series>banner"`
	FanArt     string `xml:"Series>fanart"`
	Poster     string `xml:"Series>poster"`
	Rating     string `xml:"Series>Rating"`
	FirstAired string `xml:"Series>FirstAired"`
}

type filtered_output struct {
	Title        string `json:"title"`
	Artwork      string `json:"artwork"`
	Release_date string `json:"year"`
}

// Init() must be called first to initialize the library with the two
// API keys used in it
func Init(tv_rage_api_key, tv_db_api_key string) *TVRage {
	return &TVRage{ tv_rage_api_key: tv_rage_api_key, tv_db_api_key: tv_db_api_key }
}

// Main call to get metadata for TV shows
func (tv_rage *TVRage) TVData(MediaName string) (string, error) {
	details, err := tv_rage.getSeriesDetails(MediaName)
	if err != nil {
		return "", err
	}
	tvmetadata, err := tv_rage.getTvMetadata(details)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(tvmetadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CURRENTLY UNUSED
// This call is for string correction tvdb is really good at detecting
// tv/movie titles from non-standard filenames. So we make a call to this
// function (even for movies) to get a title name in standard format that can
// then be used to query whichever online database we want without worrying
// about weird filenames. This however creates a problem if tvdb database
// doesnot have that movie/tvshow or detects it incorrectly. Tvdb always
// returns some results even if they are false. So it is hard to debug when
// tvdb errs. Sometimes when tvdb returns wrong titlename, subsequent api
// returns data for the wrong titlename
func (tv_rage *TVRage) UsableTVName(MediaName string) (string, error) {
	res, err := http.Get("http://services.tvrage.com/myfeeds/search.php?key=" + tv_rage.tv_rage_api_key + "&show=" + MediaName)
	if err != nil {
		return MediaName, err
	}
	body, err := ioutil.ReadAll(res.Body)
	var result tvrageResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return MediaName, err
	}
	if result.ShowDetails == nil {
		return MediaName, errors.New("No result obtained from tvrage for filename string correction")
	} else {

		return result.ShowDetails[0].Name, nil
	}
	return MediaName, nil
}

//get tv seriesid from tvdb using show name
func (tv_rage *TVRage) getSeriesDetails(MediaName string) (tvdbDetails, error) {
	var det tvdbDetails
	res, err := http.Get(gettvdbMirrorPath() + "api/GetSeries.php?seriesname=" + MediaName)
	if err != nil {
		return det, err
	}
	body, err := ioutil.ReadAll(res.Body)
	var results tvdbResult
	err = xml.Unmarshal(body, &results)
	if err != nil {
		return det, err
	}
	if results.SeriesDetails == nil {
		return det, errors.New("No result obtained from tvdb")
	}
	det = results.SeriesDetails[0]
	return det, nil
}

//get metadata from tvdb using seriesid
func (tv_rage *TVRage) getTvMetadata(Details tvdbDetails) (tvMetadata, error) {
	var met tvMetadata
	res, err := http.Get(gettvdbMirrorPath() + "api/" + tv_rage.tv_db_api_key + "/series/" + Details.SeriesId + "/all/" + Details.Language + ".xml")
	if err != nil {
		return met, err
	}
	body, err := ioutil.ReadAll(res.Body)
	err = xml.Unmarshal(body, &met)
	if err != nil {
		return met, err
	}
	met.Banner_Url = gettvdbMirrorPath() + "banners/"
	met.Media_type = "tv"
	return met, nil
}

// get tvdb mirrorpath - this may need change from time to time
func gettvdbMirrorPath() string {
	return "http://thetvdb.com/"
}

// convert to JSON after filtering out unwanted tv metadata
func (tv_rage *TVRage) ToJSON(data string) (string, error) {
	var f filtered_output
	var det tvMetadata
	err := json.Unmarshal([]byte(data), &det)
	if err != nil {
		return "", err
	}
	f.Title = det.SeriesName
	f.Release_date = det.FirstAired
	f.Release_date = f.Release_date[0:4]
	f.Artwork = det.Banner_Url + det.Poster

	metadata, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(metadata), nil
}
