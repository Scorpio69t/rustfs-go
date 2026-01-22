// Package bucket bucket/notification.go
package bucket

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/notification"
)

// SetNotification sets bucket event notification configuration (XML/JSON).
func (s *bucketService) SetNotification(ctx context.Context, bucketName string, config []byte) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if len(config) == 0 {
		return ErrEmptyBucketConfig
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"notification": {""}},
		ContentBody:   bytes.NewReader(config),
		ContentLength: int64(len(config)),
	}
	meta.CustomHeader.Set("Content-Type", "application/xml")

	req := core.NewRequest(ctx, http.MethodPut, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, "")
	}
	return nil
}

// GetNotification retrieves bucket event notification configuration.
func (s *bucketService) GetNotification(ctx context.Context, bucketName string) ([]byte, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"notification": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp, bucketName, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteNotification removes bucket event notification configuration.
func (s *bucketService) DeleteNotification(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"notification": {""}},
	}

	req := core.NewRequest(ctx, http.MethodDelete, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, "")
	}
	return nil
}

// ListenNotification listens for bucket events and streams notification info.
func (s *bucketService) ListenNotification(ctx context.Context, bucketName, prefix, suffix string, events []notification.EventType) <-chan notification.Info {
	notificationCh := make(chan notification.Info, 1)

	go func() {
		defer close(notificationCh)

		if err := validateBucketName(bucketName); err != nil {
			notificationCh <- notification.Info{Err: err}
			return
		}

		query := url.Values{}
		query.Set("ping", "10")
		query.Set("prefix", prefix)
		query.Set("suffix", suffix)
		for _, event := range events {
			query.Add("events", string(event))
		}

		meta := core.RequestMetadata{
			BucketName:  bucketName,
			QueryValues: query,
		}

		req := core.NewRequest(ctx, http.MethodGet, meta)
		resp, err := s.executor.Execute(ctx, req)
		if err != nil {
			notificationCh <- notification.Info{Err: err}
			return
		}
		defer closeResponse(resp)

		if resp.StatusCode != http.StatusOK {
			notificationCh <- notification.Info{Err: parseErrorResponse(resp, bucketName, "")}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		buffer := make([]byte, 0, 64*1024)
		scanner.Buffer(buffer, 4*1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}

			var info notification.Info
			if err := json.Unmarshal(line, &info); err != nil {
				notificationCh <- notification.Info{Err: err}
				return
			}

			if len(info.Records) == 0 && info.Err == nil {
				continue
			}

			select {
			case notificationCh <- info:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			notificationCh <- notification.Info{Err: err}
		}
	}()

	return notificationCh
}
