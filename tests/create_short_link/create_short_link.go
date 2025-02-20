package create_short_link

import (
	"encoding/json"
	"net/http"
	"urleater/internal/handlers"
)

func (s *createShortLinkSuite) TestCreateShortLink() {
	// 1
	longUrl1 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"

	body, code := s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "anyalias",
		LongURL:  longUrl1,
	})

	var resp1 handlers.CreateShortLinkResponse

	err := json.Unmarshal(body, &resp1)

	s.NoError(err)

	s.Equal(http.StatusOK, code)
	s.Equal(longUrl1, resp1.Link.LongUrl)

	// 2
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "any_alias",
		LongURL:  "www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 3
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "any_alias",
		LongURL:  "",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 4
	longUrl4 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias4 := "myAlias1"

	body, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: alias4,
		LongURL:  longUrl4,
	})

	var resp4 handlers.CreateShortLinkResponse

	err = json.Unmarshal(body, &resp4)

	s.NoError(err)

	s.Equal(http.StatusOK, code)
	s.Equal(longUrl4, resp4.Link.LongUrl)
	s.Assert().Equal(resp4.Link.ShortUrl, alias4)

	// 5
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "my&alias",
		LongURL:  "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 7
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "myAlias",
		LongURL:  "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 8
	longUrl8 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias8 := "myCustomAlias1234567"

	body, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: alias8,
		LongURL:  longUrl8,
	})

	var resp8 handlers.CreateShortLinkResponse

	err = json.Unmarshal(body, &resp8)

	s.NoError(err)

	s.Equal(http.StatusOK, code)
	s.Equal(longUrl8, resp8.Link.LongUrl)
	s.Assert().Equal(resp8.Link.ShortUrl, alias8)

	// 9
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "myCustomAlias12345678",
		LongURL:  "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusInternalServerError, code)

}
