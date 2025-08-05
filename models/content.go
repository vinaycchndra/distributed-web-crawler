package models

import (
	"fmt"
	"scraper/utils"
	"time"
)

type Content struct {
	Id            string    `bson:"_id,omitempty" json:"_id,omitempty"`
	Domain        string    `bson:"domain" json:"domain"`
	URL           string    `bson:"url" json:"url"`
	NormUrl       string    `bson:"norm_url" json:"norm_url"`
	Title         string    `bson:"title" json:"title"`
	Desc          string    `bson:"desc" json:"desc"`
	Author        string    `bson:"author" json:"author"`
	RawHtml       string    `bson:"raw_html" json:"raw_html"`
	Text          string    `bson:"text" json:"text"`
	DatePublished string    `bson:"date_published" json:"date_published"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

func (c *Content) InsertOne() error {
	c.Id = utils.Md5(c.URL)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	_, err := content.InsertOne(ctx, c)

	if err != nil {
		return fmt.Errorf("failed to insert doc %v, err %s", c, err.Error())
	}
	return nil
}

func InsertMany(contents []any) error {
	for _, c := range contents {
		tempConetnt := c.(*Content)
		tempConetnt.CreatedAt = time.Now()
		tempConetnt.UpdatedAt = time.Now()
	}
	_, err := content.InsertMany(ctx, contents)

	if err != nil {
		fmt.Printf("error while inserting the fetched documents err: %s\n", err.Error())
	}
	return nil
}
