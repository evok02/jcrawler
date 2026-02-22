package index

import (
	"context"
	"encoding/json"
	opensearch "github.com/opensearch-project/opensearch-go"
	"strings"
	//opensearchapi "github.com/opensearch-project/opensearch-go/opensearchapi"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"net/http"
)

const IndexName = "page_content"

var ERROR_UNSUPPORTED_DOC_TYPE = errors.New("unsupported doc type")

type Index struct {
	osClient *opensearch.Client
}

func Init(cfg *config.IndexConfig) (*Index, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{cfg.Addr},
		Username:  cfg.User,
		Password:  cfg.Pwd,
	})

	if err != nil {
		return nil, err
	}
	return &Index{osClient: client}, nil
}

func (i *Index) HandleEntry(ctx context.Context, doc any) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("HandleEntry: %s", err.Error())
	}

	switch doc := doc.(type) {
	case *db.Page:
		_, err := i.osClient.Index(
			"pages_index",
			strings.NewReader(string(data)),
			i.osClient.Index.WithDocumentID(doc.URLHash))
		if err != nil {
			return fmt.Errorf("HandleEntry: %s", err.Error())
		}
	default:
		return ERROR_UNSUPPORTED_DOC_TYPE
	}
	return nil
}
