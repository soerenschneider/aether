package weather

import (
	"errors"
	"net/http"
)

func WithHttpClient(client *http.Client) OpenweatherMapOpt {
	return func(ds *OpenweatherMapClient) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}

func WithCount(count int) OpenweatherMapOpt {
	return func(ds *OpenweatherMapClient) error {
		if count < 3 || count > 14 {
			return errors.New("count must be [3, 14]")
		}

		ds.count = count
		return nil
	}
}
