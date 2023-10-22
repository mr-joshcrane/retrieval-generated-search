package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mr-joshcrane/oracle"
	"github.com/mr-joshcrane/rag/pinecone"
)

type Corpora struct {
	id     string
	vector []float64
	text   string
}

type Rag struct {
	corpora []Corpora
	client  *pinecone.Client
	oracle  *oracle.Oracle
}

func NewRag() (*Rag, error) {
	api_key := os.Getenv("PINECONE_API_KEY")
	if api_key == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY is not set")
	}
	url := os.Getenv("PINECONE_URL")
	if url == "" {
		return nil, fmt.Errorf("PINECONE_URL is not set")
	}
	client, err := pinecone.NewClient(api_key, url)
	if err != nil {
		return nil, err
	}
	openai_api_key := os.Getenv("OPENAI_API_KEY")
	if openai_api_key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	oracle := oracle.NewOracle(openai_api_key)
	return &Rag{
		corpora: []Corpora{},
		client:  client,
		oracle:  oracle,
	}, nil
}

func (r *Rag) Answer(query string) string {
	return ""
}

func (r *Rag) AddCorpus(id string, query string) error {
	v, err := GetEmbedding(query)
	if err != nil {
		return err
	}
	resp, err := r.client.Upsert(id, v)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pinecone upsert failed: %s", resp.Status)
	}
	r.corpora = append(r.corpora, Corpora{
		vector: v,
		text:   query,
		id:     id,
	})
	return nil
}

func (r *Rag) Relevant(query string) (string, error) {
	v, err := GetEmbedding(query)
	if err != nil {
		return "", err
	}
	r.corpora = append(r.corpora, Corpora{
		vector: v,
		text:   query,
	})
	resp, err := r.client.Query(v)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var qr pinecone.QueryResponse
	if err := json.Unmarshal(data, &qr); err != nil {
		return "", err
	}
	if len(qr.Matches) > 0 {
		return qr.Matches[0].ID, nil
	}
	return "", nil
}

type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	}
}

func GetEmbedding(query string) ([]float64, error) {
	url := "https://api.openai.com/v1/embeddings"
	data, err := json.Marshal(EmbeddingRequest{
		Input: query,
		Model: "text-embedding-ada-002",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"accept":       {"application/json"},
		"content-type": {"application/json"},
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.Status)
		return nil, err
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var er EmbeddingResponse
	if err := json.Unmarshal(data, &er); err != nil {
		return nil, err
	}
	return er.Data[0].Embedding, nil
}
