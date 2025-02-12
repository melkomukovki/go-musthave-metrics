package agent

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent/config"
	pc "github.com/melkomukovki/go-musthave-metrics/internal/crypto"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	"github.com/rs/zerolog/log"
	"time"
)

type RestMetricSender struct {
	client          *resty.Client
	address         string
	config          *config.ClientConfig
	ipAddress       string
	cryptoPublicKey *rsa.PublicKey
}

func NewRestMetricSender(cfg *config.ClientConfig) (*RestMetricSender, error) {
	restSender := &RestMetricSender{
		client:  resty.New(),
		config:  cfg,
		address: cfg.Address,
	}

	restSender.client.SetRetryCount(3).
		SetRetryWaitTime(time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		})

	if cfg.CryptoKey != "" {
		publicKey, err := pc.GetPublicKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		restSender.cryptoPublicKey = publicKey
	}

	localIP, err := GetLocalIPs()
	if len(localIP) != 0 && err == nil {
		restSender.ipAddress = localIP[0]
	} else {
		log.Warn().Msg("Unable to determine local IP address")
	}

	return restSender, nil
}

func (r *RestMetricSender) SendMetrics(metrics []entities.Metric) error {
	url := fmt.Sprintf("http://%s/updates/", r.address)

	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}
	if r.ipAddress != "" {
		headers["X-Real-IP"] = r.ipAddress
	}

	mMarshaled, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error marshalling metrics: %w", err)
	}

	if r.config.HashKey != "" {
		hashString := getHashForData(mMarshaled, r.config.HashKey)
		headers["HashSHA256"] = hashString
	}

	if r.config.CryptoKey != "" {
		mMarshaled, err = pc.Encrypt(mMarshaled, r.cryptoPublicKey)
		if err != nil {
			return fmt.Errorf("error encrypting metrics: %w", err)
		}
	}

	compressedData, err := gzipData(mMarshaled)
	if err != nil {
		return fmt.Errorf("error compressing metrics: %w", err)
	}

	log.Debug().Msgf("REST headers: %+v", headers)
	resp, err := r.client.R().
		SetBody(compressedData).
		SetHeaders(headers).
		Post(url)
	if err != nil {
		return fmt.Errorf("error reporting metrics via REST: %w", err)
	}
	log.Info().Msg("Metrics were sent via REST, status: " + resp.Status())
	return nil
}
