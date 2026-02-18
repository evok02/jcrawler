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
var ERROR_EMPTY_COLLECION = errors.New("empty collection of documents")
var ERROR_UNSUCCESSFUL_TRANSACTION = errors.New("couldnt execute transaction")

type Page struct {
	URLHash       string `bson:"url_hash_id"`
	URL           string
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

func (s *Storage) GetAllPages() ([]Page, error) {
	coll := s.DB.Database("cralwer").Collection("pages")
	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	cursor, err := coll.Find(context, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("GetAllPages: %s", err)
	}

	if err := cursor.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ERROR_EMPTY_COLLECION
		}
		return nil, fmt.Errorf("GetAllPages: %s", err.Error())
	}

	var pages []Page
	if err := cursor.All(context, &pages); err != nil {
		return nil, fmt.Errorf("GetAllPages: %s", err.Error())
	}
	return pages, nil
}

func (s *Storage) InsertPage(p *Page) error {
	coll := s.DB.Database("crawler").Collection("pages")

	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	if _, err := s.GetPageByID(p.URLHash); err == nil {
		s.UpdatePageByID(p.URLHash, p)
		return nil
	}

	doc, err := bson.Marshal(p)
	if err != nil {
		return fmt.Errorf("InsertPage: %s", err.Error())
	}

	res, err := coll.InsertOne(context, doc)
	if err != nil {
		return fmt.Errorf("InsertPage: %s", err.Error())
	}

	if res.InsertedID == 0 {
		return ERROR_UNSUCCESSFUL_TRANSACTION
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
