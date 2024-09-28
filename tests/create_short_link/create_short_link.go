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
		ShortURL: "any_alias",
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
	longUrl2 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "any_alias",
		LongURL:  longUrl2,
	})

	s.Equal(http.StatusOK, code)
	s.Equal(longUrl2, resp1.Link.LongUrl)

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
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "myCustomAlias1234567",
		LongURL:  "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusOK, code)

	// 9
	_, code = s.CreateShortLink(&handlers.CreateShortLinkRequest{
		ShortURL: "myCustomAlias12345678",
		LongURL:  "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset",
	})

	s.Equal(http.StatusInternalServerError, code)

}
