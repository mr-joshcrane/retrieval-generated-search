package rag_test

import (
	"os"
	"testing"

	"github.com/mr-joshcrane/rag"
)

func init() {
	api_key := os.Getenv("PINECONE_API_KEY")
	if api_key == "" {
		panic("PINECONE_API_KEY is not set")
	}
	url := os.Getenv("PINECONE_URL")
	if url == "" {
		panic("PINECONE_URL is not set")
	}
	rag, err := rag.NewRag()
	if err != nil {
		panic(err)
	}
	err = rag.AddCorpus("1", "Red")
	if err != nil {
		panic(err)
	}
	err = rag.AddCorpus("2", "Chicken")
	if err != nil {
		panic(err)
	}
	err = rag.AddCorpus("3", "Sailor")
	if err != nil {
		panic(err)
	}
}

func TestPineconeUpsert(t *testing.T) {
	t.Parallel()
	api_key := os.Getenv("PINECONE_API_KEY")
	if api_key == "" {
		t.Fatal("PINECONE_API_KEY not set")
	}
	url := os.Getenv("PINECONE_URL")
	if url == "" {
		t.Fatal("PINECONE_URL not set")
	}
	rag, err := rag.NewRag()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		query       string
		want        string
		description string
	}{
		{
			description: "Word association, Color",
			query:       "Color",
			want:        "1",
		},
		{
			description: "Word association, Food",
			query:       "Food",
			want:        "2",
		},
		{
			description: "Word association, Occupation",
			query:       "Occupation",
			want:        "3",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			got, err := rag.Relevant(c.query)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}
