package search

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	es "github.com/elastic/go-elasticsearch/v8"
)

var ES *es.Client

func NewClientFromEnv() (*es.Client, error) {
	url := os.Getenv("ELASTIC_URL")
	if url == "" {
		return nil, fmt.Errorf("ELASTIC_URL missing")
	}

	username := os.Getenv("ELASTIC_USERNAME")
	password := os.Getenv("ELASTIC_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("ELASTIC_USERNAME or ELASTIC_PASSWORD missing")
	}

	cfg := es.Config{
		Addresses: strings.Split(url, ","),
		Username:  username,
		Password:  password,
	}

	if insecure, _ := strconv.ParseBool(os.Getenv("ELASTIC_INSECURE")); insecure {
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client, err := es.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	_, err = client.Info()
	if err != nil {
		return nil, err
	}
	ES = client
	log.Println("✅ Elasticsearch client initialized")
	return client, nil
}

func CreateIndexIfNotExists(name string, mapping string) error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	res, err := ES.Indices.Exists([]string{name})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}
	createRes, err := ES.Indices.Create(name, ES.Indices.Create.WithBody(strings.NewReader(mapping)), ES.Indices.Create.WithContext(context.Background()))
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		return fmt.Errorf("error creating index %s: %s", name, createRes.String())
	}
	log.Printf("Created index %s", name)
	return nil
}

func DeleteIndex(name string) error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	res, err := ES.Indices.Delete([]string{name}, ES.Indices.Delete.WithContext(context.Background()))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error deleting index %s: %s", name, string(body))
	}
	log.Printf("Deleted index %s", name)
	return nil
}
