package db

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

var ERROR_MULTIPLE_COPIES_OF_PAGE = errors.New("found to many entries with the same id")
var ERROR_INVALID_ID = errors.New("couldn't find entry with this id")

type Page struct {
	URLHash       string `bson:"url_hash_id"`
	Index         int
	KeywordsFound []string  `bson:"keywords_found"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

func (s *Storage) GetPageByID(id string) (*Page, error) {
	filter := bson.D{{Key: "url_hash_id", Value: id}}
	coll := s.DB.Database("crawler").Collection("pages")

	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	cursor := coll.FindOne(context, filter)
	if err := cursor.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ERROR_INVALID_ID
		}
		return nil, fmt.Errorf("GetPageByID: %s", err.Error())
	}

	var res Page
	if err := cursor.Decode(&res); err != nil {
		return nil, fmt.Errorf("GetPageByID: %s", err.Error())
	}

	return &res, nil
}

func (s *Storage) InsertPage(p *Page) error {
	coll := s.DB.Database("crawler").Collection("pages")

	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	if _, err := s.GetPageByID(p.URLHash); err == nil {
		s.UpdatePageByID(p.URLHash, p)
		return nil
	}

	if _, err := coll.InsertOne(context, p); err != nil {
		return fmt.Errorf("InsertPage: %s", err.Error())
	}

	return nil
}

func (s *Storage) DeletePageByID(id string) error {
	coll := s.DB.Database("crawler").Collection("pages")
	filter := bson.D{{Key: "url_hash_id", Value: id}}

	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	res, err := coll.DeleteMany(context, filter)
	if err != nil {
		return fmt.Errorf("DeletePageByID: %s", err.Error())
	}

	if res.DeletedCount == 0 {
		return ERROR_INVALID_ID
	}

	return nil
}

func (s *Storage) UpdatePageByID(id string, newPage *Page) (*Page, error) {
	filter := bson.D{{Key: "url_hash_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "keywords_found", Value: newPage.KeywordsFound},
		{Key: "index", Value: newPage.Index},
		{Key: "updated_at", Value: time.Now().UTC()},
	}}}
	coll := s.DB.Database("crawler").Collection("pages")

	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	res, err := coll.UpdateOne(context, filter, update)
	if err != nil {
		return nil, fmt.Errorf("ReplacePageByID: %s", err.Error())
	}

	if res.MatchedCount == 0 {
		return nil, ERROR_INVALID_ID
	}

	replaced, err := s.GetPageByID(id)
	if err != nil {
		return nil, fmt.Errorf("ReplacePageByID: %s", err.Error())
	}

	return replaced, nil
}
